package internal

import (
	"context"
	"os"
	"sync/atomic"

	"github.com/pyihe/go-pkg/syncs"
)

type stdWriter struct {
	closed int32
	wg     *syncs.WgWrapper
	ctx    context.Context
	file   *os.File
}

func newStdin(ctx context.Context, wg *syncs.WgWrapper) logWriter {
	return &stdWriter{
		ctx:  ctx,
		wg:   wg,
		file: os.Stdout,
	}
}

func (s *stdWriter) write(b []byte) (int, error) {
	if atomic.LoadInt32(&s.closed) == 1 {
		return 0, nil
	}
	return s.file.Write(b)
}

func (s *stdWriter) start() {
	s.wg.Wrap(func() {
		select {
		case <-s.ctx.Done():
			atomic.StoreInt32(&s.closed, 1)
		}
	})
}

func (s *stdWriter) stop() {

}
