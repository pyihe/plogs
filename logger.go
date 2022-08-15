package plogs

import (
	"context"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pyihe/go-pkg/buffers"
	"github.com/pyihe/go-pkg/bytes"
	"github.com/pyihe/go-pkg/syncs"
	"github.com/pyihe/plogs/internal"
	"github.com/pyihe/plogs/pkg"
)

var defaultLogger Logger

type Logger struct {
	closed int32                    // 是否关闭
	ctx    context.Context          //
	cancel context.CancelFunc       //
	waiter syncs.WgWrapper          // waiter
	once   sync.Once                // once
	writer *internal.MultipeWriters // writer
	config *LogConfig               // 配置
}

func NewLogger(opts ...Option) *Logger {
	defaultLogger.once.Do(func() {
		defaultLogger.closed = 0
		defaultLogger.ctx, defaultLogger.cancel = context.WithCancel(context.Background())
		defaultLogger.waiter = syncs.WgWrapper{}
		defaultLogger.writer = internal.NewMultipeWriters()
		defaultLogger.config = &LogConfig{
			stdout:     false,
			fileOption: WriteByLevelMerged,
			logLevel:   LevelPanic | LevelFatal | LevelError | LevelWarn | LevelInfo | LevelDebug,
			maxAge:     0,
			maxSize:    0,
			name:       "",
			logPath:    "",
		}

		for _, op := range opts {
			op(&defaultLogger)
		}

		defaultLogger.init()
		defaultLogger.start()
	})
	return &defaultLogger
}

func (l *Logger) init() {
	l.addLevelWriter()
}

func (l *Logger) Close() {
	if atomic.LoadInt32(&l.closed) == 1 {
		return
	}
	atomic.StoreInt32(&l.closed, 1)
	l.cancel()
	l.writer.Stop()
	l.waiter.Wait()
}

func (l *Logger) recover(message string) {
	if (l.config.logLevel & LevelPanic) != LevelPanic {
		return
	}
	defer func() {
		if err := recover(); err != nil {
			msg := debug.Stack()
			l.write(LevelPanic, msg)
		}
	}()

	panic(message)
}

func (l *Logger) write(level Level, message []byte) {
	config := l.config
	multipeWriter := l.writer

	if config.stdout {
		multipeWriter.WriteOne(subPath(_LevelBegin), message)
	}
	switch config.fileOption {
	case WriteByLevelMerged:
		multipeWriter.WriteOne(subPath(_LevelEnd), message)
	case WriteByLevelSeparated:
		multipeWriter.WriteOne(subPath(level), message)
	case WriteByBoth:
		multipeWriter.WriteOne(subPath(_LevelEnd), message)
		multipeWriter.WriteOne(subPath(level), message)
	}
}

func (l *Logger) log(level Level, message string) {
	_, fileName, line, ok := runtime.Caller(3)
	if !ok {
		return
	}

	var (
		appName     = l.config.name                                   // 应用名
		levelPrefix = level.prefix()                                  // 日志级别
		timeDesc    = time.Now().Format("2006/01/02 15:04:05.000000") // 时间
	)

	b := buffers.Get()
	// write app name
	if appName != "" {
		b.WriteString("[")
		b.WriteString(appName)
		b.WriteString("] ")
	}
	// write prefix
	b.WriteString(levelPrefix)

	// write timedesc
	b.WriteString("[")
	b.WriteString(timeDesc)
	b.WriteString("] ")

	// write file
	b.WriteString(fileName)
	b.WriteString(":")
	b.WriteString(strconv.FormatInt(int64(line), 10))
	b.WriteString(" ")

	// write message
	b.WriteString(message)

	//new line
	b.WriteString("\n")

	logStr := bytes.Copy(b.Bytes())
	buffers.Put(b)

	// 写入目标流
	l.write(level, logStr)
	return
}

func (l *Logger) canOutput(level Level) bool {
	if atomic.LoadInt32(&l.closed) == 1 {
		return false
	}
	if !level.valid() {
		return false
	}
	if (l.config.logLevel & level) != level {
		return false
	}
	return true
}

func (l *Logger) start() {
	if l.writer.Count() == 0 {
		panic("where the log will be written?")
	}
	l.writer.Start()
}

func (l *Logger) exit() {
	if (l.config.logLevel & LevelFatal) != LevelFatal {
		return
	}
	l.Close()
	os.Exit(1)
}

func (l *Logger) Panic(args ...interface{}) {
	if l.canOutput(LevelPanic) {
		m := pkg.GetMessage("", args)
		l.log(LevelPanic, m)
		l.recover(m)
	}
}

func (l *Logger) Panicf(template string, args ...interface{}) {
	if l.canOutput(LevelPanic) {
		m := pkg.GetMessage(template, args)
		l.log(LevelPanic, m)
		l.recover(m)
	}
}

func (l *Logger) Fatal(args ...interface{}) {
	if !l.canOutput(LevelFatal) {
		return
	}
	m := pkg.GetMessage("", args)
	l.log(LevelFatal, m)
	l.exit()
}

func (l *Logger) Fatalf(template string, args ...interface{}) {
	if !l.canOutput(LevelFatal) {
		return
	}
	m := pkg.GetMessage(template, args)
	l.log(LevelFatal, m)
	l.exit()
}

func (l *Logger) Error(args ...interface{}) {
	if !l.canOutput(LevelError) {
		return
	}
	m := pkg.GetMessage("", args)
	l.log(LevelError, m)
}

func (l *Logger) Errorf(template string, args ...interface{}) {
	if !l.canOutput(LevelError) {
		return
	}
	m := pkg.GetMessage(template, args)
	l.log(LevelError, m)
}

func (l *Logger) Warn(args ...interface{}) {
	if !l.canOutput(LevelWarn) {
		return
	}
	m := pkg.GetMessage("", args)
	l.log(LevelWarn, m)
}

func (l *Logger) Warnf(template string, args ...interface{}) {
	if !l.canOutput(LevelWarn) {
		return
	}
	m := pkg.GetMessage(template, args)
	l.log(LevelWarn, m)
}

func (l *Logger) Info(args ...interface{}) {
	if !l.canOutput(LevelInfo) {
		return
	}
	m := pkg.GetMessage("", args)
	l.log(LevelInfo, m)
}

func (l *Logger) Infof(template string, args ...interface{}) {
	if !l.canOutput(LevelInfo) {
		return
	}
	m := pkg.GetMessage(template, args)
	l.log(LevelInfo, m)
}

func (l *Logger) Debug(args ...interface{}) {
	if !l.canOutput(LevelDebug) {
		return
	}
	m := pkg.GetMessage("", args)
	l.log(LevelDebug, m)
}

func (l *Logger) Debugf(template string, args ...interface{}) {
	if !l.canOutput(LevelDebug) {
		return
	}
	m := pkg.GetMessage(template, args)
	l.log(LevelDebug, m)
}
