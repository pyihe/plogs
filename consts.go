package plogs

const (
	_LevelBegin       = iota            // begin
	LevelPanic  Level = 1 << (iota - 1) // panic
	LevelFatal                          // 致命错误, 程序会退出
	LevelError                          // 错误
	LevelWarn                           // 警告
	LevelInfo                           // 追踪
	LevelDebug                          // 调试
	_LevelEnd                           // end
)

const (
	_WriteBegin           FileOption = iota // begin
	WriteByLevelSeparated                   // 区分级别, 不同级别记录在不同目录下相应的文件中
	WriteByLevelMerged                      // 不区分级别, 所有日志记录在一个文件中
	WriteByBoth                             // 既区分级别记录也一起记录
	_WriteEnd                               // end
)

type (
	Level      int // Level 日志级别
	FileOption int // FileOption 日志文件写选项
)

func (w FileOption) valid() bool {
	return w < _WriteEnd && w > _WriteBegin
}

func (l Level) valid() bool {
	return l > _LevelBegin && l < _LevelEnd
}

// 每个级别的日志对应的prefix
func (l Level) prefix() (prefix string) {
	switch l {
	case LevelPanic:
		prefix = "[P] "
	case LevelFatal:
		prefix = "[F] "
	case LevelError:
		prefix = "[E] "
	case LevelWarn:
		prefix = "[W] "
	case LevelInfo:
		prefix = "[I] "
	case LevelDebug:
		prefix = "[D] "
	}
	return
}

// 需要区分级别存放日志信息时，用于获取每个级别日志存放的子目录
func (l Level) subPath() (suffix string) {
	switch l {
	case LevelPanic:
		suffix = "panics"
	case LevelFatal:
		suffix = "fatals"
	case LevelError:
		suffix = "errors"
	case LevelWarn:
		suffix = "warns"
	case LevelInfo:
		suffix = "infos"
	case LevelDebug:
		suffix = "debugs"
	case _LevelEnd:
		suffix = "merged"
	}
	return
}
