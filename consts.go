package plogs

import "time"

const (
	defaultCutDuration = 1 * time.Second  // 默认日志切割周期
	defaultFileSize    = 10 * 1024 * 1024 // 默认文件大小60MB
)

const (
	_LevelBegin        = iota            // begin
	LevelPanic   Level = 1 << (iota - 1) // panic
	LevelFatal                           // 致命错误, 程序会退出
	LevelError                           // 错误
	LevelWarning                         // 警告
	LevelInfo                            // 追踪
	LevelDebug                           // 调试
	_LevelEnd                            // end
)

const (
	_CutBegin     CutOption = iota // begin
	CutDaily                       // 根据时间周期，每天切割
	CutHourly                      // 根据时间周期，每小时切割
	CutHalfAnHour                  // 根据时间周期，每半小时切割
	CutTenMin                      // 根据时间周期，每10分钟切割
	CutPer10M                      // 根据文件大小，每10M切割一次
	CutPer60M                      // 根据文件大小，每60M切割一次
	CutPer100M                     // 根据文件大小，每100M切割一次
	_CutEnd                        // end
)

const (
	_WriteBegin   WriteOption = iota // begin
	WriteByLevel                     // 区分级别, 不同级别记录在不同目录下相应的文件中
	WriteByMerged                    // 不区分级别, 所有日志记录在一个文件中
	WriteByAll                       // 既区分级别记录也一起记录
	_WriteEnd                        // end
)

type (
	CutOption   int // CutOption 日志文件切割周期选项
	Level       int // Level 日志级别
	WriteOption int // WriteOption 日志写选项
)

func (w WriteOption) valid() bool {
	return w < _WriteEnd && w > _WriteBegin
}

func (c CutOption) valid() bool {
	return c > _CutBegin && c < _CutEnd
}

func (l Level) valid() bool {
	return l > _LevelBegin && l < _LevelEnd
}

// 每个级别的日志对应的prefix
func (l Level) prefix() (prefix string) {
	switch l {
	case LevelPanic:
		prefix = "[Panic]"
	case LevelFatal:
		prefix = "[Fatal]"
	case LevelError:
		prefix = "[Error]"
	case LevelWarning:
		prefix = "[Warn] "
	case LevelInfo:
		prefix = "[Info] "
	case LevelDebug:
		prefix = "[Debug]"
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
	case LevelWarning:
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

func (l Level) colorCode() (code int) {
	switch l {
	case LevelPanic: // 红色
		code = 31
	case LevelFatal: // 红色
		code = 31
	case LevelError: // 紫红色
		code = 35
	case LevelWarning: // 黄色
		code = 33
	case LevelInfo: // 绿色
		code = 34
	case LevelDebug: // 白色
	}
	return
}
