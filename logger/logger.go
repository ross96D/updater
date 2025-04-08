package logger

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ross96D/updater/share/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type message []byte

// small and simple implementation of a CircularQueue
// not thread safe
type CircularQueue struct {
	// if queue gets filled incoming messages will be dropped
	messages [256]message
	head     uint8
	tail     uint8

	mut sync.RWMutex
}

func (q *CircularQueue) InQueue() uint8 {
	q.mut.RLock()
	// uses uint8 overflow
	r := q.tail - q.head
	q.mut.RUnlock()
	return r
}

func (q *CircularQueue) Space() uint8 {
	q.mut.RLock()
	// uses uint8 overflow
	r := (q.head - 1) - q.tail
	q.mut.RUnlock()
	return r

}

func (q *CircularQueue) push(m message) {
	if q.Space() == 0 {
		return
	}

	q.mut.Lock()
	// copy data as the incoming message can become corrupted
	var mnew message
	if q.messages[q.tail] != nil && cap(q.messages[q.tail]) >= len(m) {
		mnew = q.messages[q.tail][:len(m)]
	} else {
		mnew = make(message, len(m))
	}

	copy(mnew, m)
	q.messages[q.tail] = mnew
	q.tail++
	q.mut.Unlock()
}

func (q *CircularQueue) pop() (message, bool) {
	if q.InQueue() == 0 {
		return nil, false
	}
	q.mut.Lock()
	result := q.messages[q.head]
	q.head++
	q.mut.Unlock()
	return result, true
}

func (q *CircularQueue) Write(p []byte) (int, error) {
	q.push(message(p))
	return len(p), nil
}

type fileWriter struct {
	file     *os.File
	filepath string
}

func (w *fileWriter) write(m message) error {
	_, err := w.file.Write(m)
	if err != nil {
		return err
	}
	// TODO Should we sync here????
	// Maybe is necessary for showing up to date content on the web
	return w.file.Sync()
}

type responseWriter struct {
	ctx        context.Context
	controller *http.ResponseController
	writer     http.ResponseWriter
}

type CtxExpired string

func (CtxExpired) Error() string {
	return "context expired"
}

const errCtxExpired CtxExpired = ""

func (w *responseWriter) write(m message) error {
	if w.ctx.Err() != nil {
		return errCtxExpired
	}
	_, err := w.writer.Write(m)
	if err != nil {
		return err
	}
	return w.controller.Flush()
}

type Handler struct {
	resp  *responseWriter
	file  fileWriter
	queue *CircularQueue

	sendSignal chan struct{}
	endSignal  chan struct{}

	isHandling atomic.Bool
}

func (h *Handler) FileName() string {
	return filepath.Base(h.file.file.Name())
}

func (h *Handler) Write(p []byte) (int, error) {
	return h.queue.Write(p)
}

func NewHandler(w http.ResponseWriter, r *http.Request) *Handler {
	_, file, err := utils.CreateTempFile()
	utils.Assert(err == nil, "NewHandler failed to create temp file %s", err)

	sendSignal := make(chan struct{})
	endSignal := make(chan struct{})

	return &Handler{
		sendSignal: sendSignal,
		endSignal:  endSignal,

		queue: &CircularQueue{},

		resp: &responseWriter{
			ctx:        r.Context(),
			writer:     w,
			controller: http.NewResponseController(w),
		},
		file: fileWriter{
			file: file,
		},
	}
}

func (h *Handler) write(m message) error {
	var err1 error
	if h.resp != nil {
		err1 = h.resp.write(m)
		if err1 == errCtxExpired {
			err1 = nil
			h.resp = nil
		}
	}
	err2 := h.file.write(m)
	return errors.Join(err1, err2)
}

func (h *Handler) sendAll() {
	for {
		m, ok := h.queue.pop()
		if !ok {
			break
		}
		if err := h.write(m); err != nil {
			log.Debug().Err(err).Caller().Send()
			break
		}
	}
}

func (h *Handler) Handle() {
	h.isHandling.Store(true)
	for {
		select {
		case <-h.endSignal:
			goto end
		case <-h.sendSignal:
			h.sendAll()
			// notify the sendAll finished
			h.sendSignal <- struct{}{}
		default:
			m, ok := h.queue.pop()
			if !ok {
				// allow context switch and reduce cpu pressure
				time.Sleep(100 * time.Microsecond)
				continue
			}
			if err := h.write(m); err != nil {
				log.Debug().Err(err).Caller().Send()
				goto end
			}
		}
	}
end:
	h.isHandling.Store(false)
}

func (h *Handler) SendAll(timeout *time.Timer) {
	// guard invalid calls to channels
	if !h.isHandling.Load() {
		return
	}
	h.sendSignal <- struct{}{}
	select {
	case <-h.sendSignal:
	case <-timeout.C:
	}
}

func (h *Handler) End() {
	// guard invalid calls to channels
	if !h.isHandling.Load() {
		return
	}
	h.sendSignal <- struct{}{}
	t := time.NewTimer(100 * time.Millisecond)
	select {
	case <-h.sendSignal:
	case <-t.C:
		go func() {
			h.endSignal <- struct{}{}
		}()
	}
}

func New(logger zerolog.Logger, r *http.Request, w http.ResponseWriter) (handler *Handler, l *zerolog.Logger, err error) {
	handler = NewHandler(w, r)
	wait := atomic.Bool{}
	wait.Store(true)
	go func() {
		wait.Store(false)
		handler.Handle()
	}()
	// allow the handler to be run first
	for wait.Load() {
		runtime.Gosched()
	}
	writer := zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.Out = handler
	})
	logger = logger.Output(writer)
	l = &logger
	return
}
