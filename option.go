package plogs

import (
	"time"

	"github.com/pyihe/plogs/internal"
)

type Option func(c *Logger)

// LogConfig 配置项
type LogConfig struct {
	stdout     bool          // 是否stdin输出
	fileOption FileOption    // 日志记录方式
	logLevel   Level         // 需要记录的日志级别
	maxAge     time.Duration // 日志文件保存最长时间
	maxSize    int64         // 日志文件大小上限
	name       string        // 日志来自哪个应用
	logPath    string        // 日志存储路径
}

// WithStdout 设置是否同步输出到标准输出
func WithStdout(b bool) Option {
	return func(c *Logger) {
		c.config.stdout = b
	}
}

// WithFileOption 设置写文件选项
func WithFileOption(opt FileOption) Option {
	return func(c *Logger) {
		if opt.valid() {
			c.config.fileOption = opt
		}
	}
}

// WithLogLevel 日志记录级别: [ LevelFatal | LevelFatal | LevelError | LevelWarn | LevelInfo | LevelDebug ]
func WithLogLevel(level Level) Option {
	return func(c *Logger) {
		c.config.logLevel = level
	}
}

// WithMaxAge 设置日志文件保存最长时间
func WithMaxAge(t time.Duration) Option {
	return func(c *Logger) {
		c.config.maxAge = t
	}
}

// WithMaxSize 设置日志文件保存上限
func WithMaxSize(size int64) Option {
	return func(c *Logger) {
		c.config.maxSize = size
	}
}

// WithName 设置app名称
func WithName(name string) Option {
	return func(c *Logger) {
		c.config.name = name
	}
}

// WithLogPath 设置日志文件存放目录(如果区分级别存放日志，将会在filepath下创建对应级别的目录用于区分)
func WithLogPath(filepath string) Option {
	return func(c *Logger) {
		if filepath != "" {
			c.config.logPath = filepath
		}
	}
}

// WithWriter 添加自定义Writer
func WithWriter(writer ...internal.LogWriter) Option {
	return func(c *Logger) {
		for _, w := range writer {
			c.writer.AddWriter(w)
		}
	}
}
