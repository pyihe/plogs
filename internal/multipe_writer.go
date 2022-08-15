package internal

import (
	"github.com/pyihe/go-pkg/errors"
	"github.com/pyihe/go-pkg/strings"
)

type LogWriter interface {
	Write(b []byte) (int, error)
	Name() string
	Start()
	Stop()
}

type MultipeWriters struct {
	writers map[string]LogWriter // writers
}

func NewMultipeWriters() *MultipeWriters {
	mw := &MultipeWriters{
		writers: make(map[string]LogWriter),
	}
	return mw
}

func (m *MultipeWriters) AddWriter(writer LogWriter) {
	if writer == nil {
		return
	}
	m.writers[strings.ToLower(writer.Name())] = writer
	return
}

func (m *MultipeWriters) Write(b []byte) (n int, err error) {
	for _, w := range m.writers {
		n, err = w.Write(b)
	}
	return
}

func (m *MultipeWriters) WriteOne(name string, b []byte) (n int, err error) {
	writer := m.writers[strings.ToLower(name)]
	if writer != nil {
		return writer.Write(b)
	}
	return 0, errors.New("not found writer")
}

func (m *MultipeWriters) Start() {
	for _, w := range m.writers {
		w.Start()
	}
}

func (m *MultipeWriters) Stop() {
	for _, w := range m.writers {
		w.Stop()
	}
}

func (m *MultipeWriters) Count() (n int) {
	n = len(m.writers)
	return
}
