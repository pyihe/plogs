package plogs

import "time"

type Option func(c *LogConfig)

// LogConfig 配置项
type LogConfig struct {
	stdout        bool          // 是否同时输出到stdout
	bufferSize    int           // 日志通道缓冲区长度
	writeOption   WriteOption   // 日志记录方式
	cutOption     CutOption     // 日志切割方式
	writeLevel    Level         // 需要记录的日志级别
	fileLimit     int           // 文件数量限制
	fileMaxTime   int64         // 单个文件最长保存期限
	flushDuration time.Duration // 日志通道缓冲区flush周期, 单位毫秒
	appName       string        // 日志来自哪个应用
	logPath       string        // 日志存储路径
}

// WithAppName 设置app名称
func WithAppName(name string) Option {
	return func(c *LogConfig) {
		if len(name) > 0 {
			c.appName = name
		}
	}
}

// WithWriteOption 设置写文件选项
func WithWriteOption(opt WriteOption) Option {
	return func(c *LogConfig) {
		if opt.valid() {
			c.writeOption = opt
		}
	}
}

// WithCutOption 日志文件切割方式
func WithCutOption(cutOption CutOption) Option {
	return func(c *LogConfig) {
		if cutOption.valid() {
			c.cutOption = cutOption
		}
	}
}

// WithWriteLevel 日志记录级别: [ LevelFatal, LevelFatal | LevelError | LevelWarning | LevelInfo | LevelDebug ]
func WithWriteLevel(level Level) Option {
	return func(c *LogConfig) {
		c.writeLevel = level
	}
}

// WithStdout 设置是否在终端显示日志输出
func WithStdout(b bool) Option {
	return func(c *LogConfig) {
		c.stdout = b
	}
}

// WithLogPath 设置日志文件存放目录(如果区分级别存放日志，将会在filepath下创建对应级别的目录用于区分)
func WithLogPath(filepath string) Option {
	return func(c *LogConfig) {
		if filepath != "" {
			c.logPath = filepath
		}
	}
}

// WithBufferSize 设置日志通道buffer大小: [1024, 1024000]
func WithBufferSize(size int) Option {
	return func(c *LogConfig) {
		if size < 1024 {
			size = 1024
		}
		if size > 1024000 {
			size = 1024000
		}
		c.bufferSize = size
	}
}

// WithFlushDuration 多久刷盘一次, 单位毫秒: [500ms-5000ms]
func WithFlushDuration(duration time.Duration) Option {
	return func(c *LogConfig) {
		if duration < 500*time.Millisecond {
			duration = 500
		}
		if duration > 5000*time.Millisecond {
			duration = 5000
		}
		c.flushDuration = duration
	}
}

// WithMaxLimit 设置日志文件最大保存数量
func WithMaxLimit(n int) Option {
	return func(c *LogConfig) {
		c.fileLimit = n
	}
}

// WithMaxTime 设置日志文件最大保存时间, 单位: 秒
func WithMaxTime(sec int64) Option {
	return func(c *LogConfig) {
		c.fileMaxTime = sec
	}
}
