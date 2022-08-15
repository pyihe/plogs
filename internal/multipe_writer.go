package internal

import (
	"context"
	"time"

	"github.com/pyihe/go-pkg/syncs"
)

type logWriter interface {
	write(b []byte) (int, error)
	start()
	stop()
}

type MultipeWriters struct {
	wg      syncs.WgWrapper   // waiter
	writers map[int]logWriter // writers
}

func NewMultipeWriters() *MultipeWriters {
	mw := &MultipeWriters{
		wg:      syncs.WgWrapper{},
		writers: make(map[int]logWriter),
	}
	return mw
}

func (m *MultipeWriters) AddFileWriter(ctx context.Context, k int, filePath, fileName string, maxSize int64, maxAge time.Duration) {
	writer := newFileWriter(ctx, &m.wg, filePath, fileName, maxSize, maxAge)
	m.writers[k] = writer
}

func (m *MultipeWriters) AddStdWriter(ctx context.Context, k int) {
	m.writers[k] = newStdin(ctx, &m.wg)
}

func (m *MultipeWriters) Write(k int, b []byte) {
	writer := m.writers[k]
	if writer != nil {
		writer.write(b)
	}
}

func (m *MultipeWriters) Start() {
	for _, w := range m.writers {
		w.start()
	}
}

func (m *MultipeWriters) Stop() {
	for _, w := range m.writers {
		w.stop()
	}
	m.wg.Wait()
}

func (m *MultipeWriters) Count() int {
	return len(m.writers)
}
