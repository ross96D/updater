package match

import (
	"bytes"
	"errors"
	"io"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ross96D/updater/share/configuration"
	"github.com/ross96D/updater/share/utils"
	"github.com/rs/zerolog"
)

type pipeType int

const (
	stdout pipeType = iota + 1
	stderr
)

type streamConsumer struct {
	logger      *zerolog.Logger
	ptype       pipeType
	notConsumed []byte
	isClosed    atomic.Bool
	mut         sync.Mutex
}

func (consumer *streamConsumer) append(p []byte) {
	consumer.mut.Lock()
	defer consumer.mut.Unlock()

	if consumer.notConsumed == nil {
		consumer.notConsumed = make([]byte, 0, len(p))
	}
	consumer.notConsumed = append(consumer.notConsumed, p...)
}

func (consumer *streamConsumer) consume(all bool) {
	consumer.mut.Lock()
	defer consumer.mut.Unlock()

	var logger *zerolog.Event
	switch consumer.ptype {
	case stdout:
		logger = consumer.logger.Info().Str("pipe", "stdout")
	case stderr:
		logger = consumer.logger.Warn().Str("pipe", "stderr")
	}

	start := 0
	for start < len(consumer.notConsumed) {
		index := bytes.IndexByte(consumer.notConsumed[start:], '\n')
		if index == -1 {
			break
		}
		// TODO test what happens if index + 1 equals len
		logger.Msg(string(consumer.notConsumed[start : index+1]))
		start = index + 1
	}
	copy(consumer.notConsumed, consumer.notConsumed[start:])
	consumer.notConsumed = consumer.notConsumed[:len(consumer.notConsumed)-start]

	if all && len(consumer.notConsumed) > 0 {
		logger.Msg(string(consumer.notConsumed))
		consumer.notConsumed = consumer.notConsumed[0:0]
	}
}

func (consumer *streamConsumer) Write(p []byte) (int, error) {
	if consumer.isClosed.Load() {
		return 0, errors.New("called write on closed stream")
	}
	if len(p) == 0 {
		return 0, nil
	}
	consumer.append(p)
	consumer.consume(false)

	return len(p), nil
}

func RunCommand(logger *zerolog.Logger, command configuration.Command) error {
	cmd := exec.Command(command.Command, command.Args...)
	if command.Path != "" {
		cmd.Path = command.Path
	}

	buffout := &utils.StreamBuffer{}
	cmd.Stdout = buffout
	bufferr := &utils.StreamBuffer{}
	cmd.Stderr = bufferr

	outconsumer := &streamConsumer{logger: logger, ptype: stdout}
	errconsumer := &streamConsumer{logger: logger, ptype: stdout}
	//nolint errcheck
	go io.Copy(outconsumer, buffout)
	//nolint errcheck
	go io.Copy(errconsumer, bufferr)

	cmd.WaitDelay = time.Millisecond
	err := cmd.Run()

	outconsumer.isClosed.Store(true)
	errconsumer.isClosed.Store(true)
	outconsumer.consume(true)
	errconsumer.consume(true)

	if err != nil {
		logger.Error().Err(err).Msgf("post command %s", cmd.String())
		return ErrError{err}
	}
	return nil
}