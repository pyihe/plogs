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
	"github.com/pyihe/plogs/internal"
	"github.com/pyihe/plogs/pkg"
)

var defaultLogger *Logger

type Logger struct {
	closed int32              // 是否关闭
	ctx    context.Context    //
	cancel context.CancelFunc //
	once   sync.Once          // once
	pool   sync.Pool
	writer *internal.MultipeWriters // writer
	config *LogConfig               // 配置
}

func NewLogger(opts ...Option) *Logger {
	if defaultLogger != nil {
		return defaultLogger
	}

	defaultLogger = &Logger{
		once:   sync.Once{},
		writer: internal.NewMultipeWriters(),
		config: &LogConfig{
			stdout:     false,
			fileOption: WriteByLevelMerged,
			logLevel:   LevelPanic | LevelFatal | LevelError | LevelWarn | LevelInfo | LevelDebug,
			maxAge:     0,
			maxSize:    0,
			name:       "", // 默认没有prefix tag
			logPath:    "", // 默认不存放日志文件
		},
	}

	defaultLogger.ctx, defaultLogger.cancel = context.WithCancel(context.Background())

	defaultLogger.once.Do(func() {
		for _, op := range opts {
			op(defaultLogger.config)
		}
	})

	defaultLogger.init()
	defaultLogger.start()

	return defaultLogger
}

func (l *Logger) init() {
	config := l.config
	// 添加标准输出流
	if config.stdout {
		l.writer.AddStdWriter(l.ctx, _LevelBegin)
	}
	// 如果没有指定日志存储目录，则不需要记录到文件中
	if config.logPath == "" {
		return
	}
	allLevels := []Level{
		LevelPanic, LevelFatal, LevelError, LevelWarn, LevelInfo, LevelDebug,
	}
	// 添加文件输出流
	switch config.fileOption {
	case WriteByLevelMerged: // 归并的话，只需要将所有日志文件记录在一个文件中
		l.writer.AddFileWriter(l.ctx, int(_LevelEnd), config.logPath, "temp.log", config.maxSize, config.maxAge)
	case WriteByLevelSeparated: // 每个级别的日志分别记录在自己的文件中
		for _, lv := range allLevels {
			if l.outputLevel(lv) {
				levelPath := pkg.JoinPath(config.logPath, lv.subPath())
				l.writer.AddFileWriter(l.ctx, int(lv), levelPath, "temp.log", config.maxSize, config.maxAge)
			}
		}
	case WriteByBoth: // 上述两者同时存在
		targetLevel := make([]Level, 0, 8)
		targetLevel = append(targetLevel, _LevelEnd)
		for _, lv := range allLevels {
			if l.outputLevel(lv) {
				targetLevel = append(targetLevel, lv)
			}
		}
		for _, lv := range targetLevel {
			levelPath := pkg.JoinPath(config.logPath, lv.subPath())
			l.writer.AddFileWriter(l.ctx, int(lv), levelPath, "temp.log", config.maxSize, config.maxAge)
		}
	}
}

func (l *Logger) Close() {
	if atomic.LoadInt32(&l.closed) == 1 {
		return
	}
	atomic.StoreInt32(&l.closed, 1)
	l.cancel()
	l.writer.Stop()
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
	if config.stdout {
		l.writer.Write(_LevelBegin, message)
	}
	switch config.fileOption {
	case WriteByLevelMerged:
		l.writer.Write(int(_LevelEnd), message)
	case WriteByLevelSeparated:
		l.writer.Write(int(level), message)
	case WriteByBoth:
		l.writer.Write(int(level), message)
		l.writer.Write(int(_LevelEnd), message)
	}
}

func (l *Logger) log(level Level, message string) {
	if atomic.LoadInt32(&l.closed) == 1 || !level.valid() {
		return
	}
	if l.outputLevel(level) == false {
		return
	}

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

func (l *Logger) outputLevel(level Level) bool {
	return (l.config.logLevel & level) == level
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

func (l *Logger) panic(args ...interface{}) {
	m := getMessage("", args)
	l.log(LevelPanic, m)
	l.recover(m)
}

func (l *Logger) panicf(template string, args ...interface{}) {
	m := getMessage(template, args)
	l.log(LevelPanic, m)
	l.recover(m)
}

func (l *Logger) fatal(args ...interface{}) {
	m := getMessage("", args)
	l.log(LevelFatal, m)
	l.exit()
}

func (l *Logger) fatalf(template string, args ...interface{}) {
	m := getMessage(template, args)
	l.log(LevelFatal, m)
	l.exit()
}

func (l *Logger) error(args ...interface{}) {
	m := getMessage("", args)
	l.log(LevelError, m)
}

func (l *Logger) errorf(template string, args ...interface{}) {
	m := getMessage(template, args)
	l.log(LevelError, m)
}

func (l *Logger) Warn(args ...interface{}) {
	m := getMessage("", args)
	l.log(LevelWarn, m)
}

func (l *Logger) Warnf(template string, args ...interface{}) {
	m := getMessage(template, args)
	l.log(LevelWarn, m)
}

func (l *Logger) Info(args ...interface{}) {
	m := getMessage("", args)
	l.log(LevelInfo, m)
}

func (l *Logger) Infof(template string, args ...interface{}) {
	m := getMessage(template, args)
	l.log(LevelInfo, m)
}

func (l *Logger) Debug(args ...interface{}) {
	m := getMessage("", args)
	l.log(LevelDebug, m)
}

func (l *Logger) Debugf(template string, args ...interface{}) {
	m := getMessage(template, args)
	l.log(LevelDebug, m)
}
