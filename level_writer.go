package plogs

import (
	"github.com/pyihe/plogs/internal"
	"github.com/pyihe/plogs/pkg"
)

type levelWriter struct {
	internal.LogWriter
	level Level
}

func (l *Logger) addLevelWriter() {
	config := l.config
	if config.stdout {
		writer := &levelWriter{
			level: _LevelBegin,
		}
		writer.LogWriter = internal.NewStdWriter(l.ctx, &l.waiter)
		l.writer.AddWriter(writer)
	}
	if config.logPath == "" {
		return
	}
	allLevels := []Level{
		LevelPanic, LevelFatal, LevelError, LevelWarn, LevelInfo, LevelDebug,
	}
	switch config.fileOption {
	case WriteByLevelMerged:
		writer := &levelWriter{
			LogWriter: internal.NewFileWriter(l.ctx, &l.waiter, config.logPath, "temp.log", config.maxSize, config.maxAge),
			level:     _LevelEnd,
		}
		l.writer.AddWriter(writer)
	case WriteByLevelSeparated:
		for _, level := range allLevels {
			if l.outputLevel(level) {
				writer := &levelWriter{
					LogWriter: internal.NewFileWriter(l.ctx, &l.waiter, pkg.JoinPath(config.logPath, subPath(level)), "temp.log", config.maxSize, config.maxAge),
					level:     level,
				}
				l.writer.AddWriter(writer)
			}
		}
	case WriteByBoth:
		targetLevel := make([]Level, 0, 8)
		targetLevel = append(targetLevel, _LevelEnd)
		for _, lv := range allLevels {
			if l.outputLevel(lv) {
				targetLevel = append(targetLevel, lv)
			}
		}
		for _, level := range targetLevel {
			writer := &levelWriter{
				LogWriter: internal.NewFileWriter(l.ctx, &l.waiter, pkg.JoinPath(config.logPath, subPath(level)), "temp.log", config.maxSize, config.maxAge),
				level:     level,
			}
			l.writer.AddWriter(writer)
		}
	}
}

func (lw *levelWriter) Name() string {
	return subPath(lw.level)
}
