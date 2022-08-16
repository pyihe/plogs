package internal

import (
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

func (m *MultipeWriters) WriteTo(b []byte, names ...string) (n int, err error) {
	for _, name := range names {
		writer, exist := m.writers[strings.ToLower(name)]
		if exist {
			n, err = writer.Write(b)
		}
	}
	return
}

func (m *MultipeWriters) Write(b []byte) (n int, err error) {
	for _, w := range m.writers {
		n, err = w.Write(b)
	}
	return
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
