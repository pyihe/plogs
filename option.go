package plogs

import "time"

type Option func(logger *Logger)

// WithAppName 设置app名称
func WithAppName(name string) Option {
	return func(logger *Logger) {
		if len(name) > 0 {
			logger.appName = name
		}
	}
}

// WithWriteOption 设置写文件选项
func WithWriteOption(opt WriteOption) Option {
	return func(logger *Logger) {
		if opt.valid() {
			logger.writeOption = opt
		}
	}
}

// WithCutOption 日志文件切割方式
func WithCutOption(cutOption CutOption) Option {
	return func(logger *Logger) {
		if cutOption.valid() {
			logger.cutOption = cutOption
		}
	}
}

// WithWriteLevel 日志记录级别: [ LevelFatal, LevelFatal | LevelError | LevelWarning | LevelInfo | LevelDebug ]
func WithWriteLevel(level Level) Option {
	return func(logger *Logger) {
		logger.writeLevel = level
	}
}

// WithStdout 设置是否在终端显示日志输出
func WithStdout(b bool) Option {
	return func(logger *Logger) {
		logger.stdTag = b
	}
}

// WithLogPath 设置日志文件存放目录(如果区分级别存放日志，将会在filepath下创建对应级别的目录用于区分)
func WithLogPath(filepath string) Option {
	return func(logger *Logger) {
		if filepath != "" {
			logger.logSavePath = filepath
		}
	}
}

// WithBufferSize 设置日志通道buffer大小: [1024, 1024000]
func WithBufferSize(size int) Option {
	return func(logger *Logger) {
		if size < 1024 {
			size = 1024
		}
		if size > 1024000 {
			size = 1024000
		}
		logger.msgChanBufferSize = size
		logger.msgChan = make(chan *logMessage, size)
	}
}

// WithFlushDuration 多久刷盘一次, 单位毫秒: [500ms-5000ms]
func WithFlushDuration(duration time.Duration) Option {
	return func(logger *Logger) {
		if duration < 500*time.Millisecond {
			duration = 500
		}
		if duration > 5000*time.Millisecond {
			duration = 5000
		}
		logger.flushDuration = duration
	}
}
