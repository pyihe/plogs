package internal

import (
	"context"
	"os"
	"sync/atomic"

	"github.com/pyihe/go-pkg/syncs"
)

type stdWriter struct {
	closed      int32
	wg          *syncs.WgWrapper
	ctx         context.Context
	writeBuffer chan []byte
}

func NewStdWriter(ctx context.Context, wg *syncs.WgWrapper) LogWriter {
	return &stdWriter{
		ctx:         ctx,
		wg:          wg,
		writeBuffer: make(chan []byte, bufferSize),
	}
}

func (s *stdWriter) Name() string {
	return "stdout"
}

func (s *stdWriter) Write(b []byte) (int, error) {
	if atomic.LoadInt32(&s.closed) == 1 {
		return 0, nil
	}
	s.writeBuffer <- b
	return len(b), nil
}

func (s *stdWriter) Start() {
	s.wg.Wrap(func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			case msg := <-s.writeBuffer:
				os.Stdout.Write(msg)
			}
		}
	})
}

func (s *stdWriter) Stop() {
	if atomic.LoadInt32(&s.closed) == 1 {
		return
	}
	atomic.StoreInt32(&s.closed, 1)
	s.clean()
	close(s.writeBuffer)
}

func (s *stdWriter) clean() {
	if count := len(s.writeBuffer); count > 0 {
		remainMsg := make([][]byte, count)
		index := 0
		for msg := range s.writeBuffer {
			remainMsg[index] = msg
			index++
			if index == count {
				break
			}
		}
		for _, m := range remainMsg {
			os.Stdout.Write(m)
		}
	}
}
