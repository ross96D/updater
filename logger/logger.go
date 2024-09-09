package logger

import (
	"context"
	"io"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type message []byte

// small and simple implementation of a queue
// not thread safe
type queue struct {
	// if queue gets filled incoming messages will be dropped
	messages [256]message
	head     uint8
	tail     uint8
}

func (q *queue) InQueue() uint8 {
	// uses uint8 overflow
	return q.tail - q.head
}

func (q *queue) Space() uint8 {
	// uses uint8 overflow
	return (q.head - 1) - q.tail
}

func (q *queue) push(m message) {
	if q.Space() == 0 {
		return
	}
	q.messages[q.tail] = m
	q.tail++
}

func (q *queue) pop() (message, bool) {
	if q.InQueue() == 0 {
		return nil, false
	}
	result := q.messages[q.head]
	q.head++
	return result, true
}

func (q *queue) Write(p []byte) (int, error) {
	q.push(message(p))
	return len(p), nil
}

type responseHandler struct {
	ctx        context.Context
	controller *http.ResponseController
	writer     http.ResponseWriter
	q          *queue
	endCh      chan bool
	notify     chan bool
}

func (handler *responseHandler) write(m message) error {
	_, err := handler.writer.Write(m)
	if err != nil {
		return err
	}
	return handler.controller.Flush()
}

func (handler *responseHandler) sendAll() {
	for {
		m, ok := handler.q.pop()
		if !ok {
			break
		}
		if err := handler.write(m); err != nil {
			log.Debug().Err(err).Caller().Send()
			break
		}
	}
}

func (handler *responseHandler) handle() {
	handler.endCh = make(chan bool)
	for {
		select {
		case <-handler.ctx.Done():
			handler.sendAll()
			goto end
		case <-handler.endCh:
			handler.sendAll()
			goto end
		default:
			m, ok := handler.q.pop()
			if !ok {
				// allow context switch and reduce cpu pressure
				time.Sleep(100 * time.Microsecond)
				continue
			}
			if err := handler.write(m); err != nil {
				log.Debug().Err(err).Caller().Send()
				goto end
			}
		}
	}
end:
	if handler.notify != nil {
		handler.notify <- true
	}
}

func (handler *responseHandler) end() {
	handler.notify = make(chan bool)
	handler.endCh <- true
	t := time.NewTimer(100 * time.Millisecond)
	select {
	case <-handler.notify:
	case <-t.C:
	}
}

func New(ctx context.Context, logger zerolog.Logger, responseWriter http.ResponseWriter, logWriter io.Writer) (l zerolog.Logger, end func()) {
	queue := &queue{}
	handler := &responseHandler{
		controller: http.NewResponseController(responseWriter),
		writer:     responseWriter,
		q:          queue,
		ctx:        ctx,
	}
	wait := atomic.Bool{}
	wait.Store(true)
	go func() {
		wait.Store(false)
		handler.handle()
	}()
	// allow the handler to be run first
	for wait.Load() {
		runtime.Gosched()
	}
	writer := zerolog.MultiLevelWriter(
		zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
			w.Out = queue
		}),
		logWriter,
	)

	return logger.Output(writer), handler.end
}
