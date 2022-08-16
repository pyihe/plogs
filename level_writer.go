package plogs

import (
	"github.com/pyihe/plogs/internal"
	"github.com/pyihe/plogs/pkg"
)

type levelWriter struct {
	internal.LogWriter
	level Level
}

func (l *Logger) addLevelWriter() error {
	var err error
	var config = l.config
	var allLevels = []Level{
		LevelPanic, LevelFatal, LevelError, LevelWarn, LevelInfo, LevelDebug,
	}

	if config.stdout {
		writer := &levelWriter{
			level: _LevelBegin,
		}
		writer.LogWriter, _ = internal.NewStdWriter(l.ctx, &l.waiter)
		l.writer.AddWriter(writer)
	}

	if config.logPath == "" {
		return nil
	}

	switch config.fileOption {
	case WriteByLevelMerged:
		writer := &levelWriter{
			level: _LevelEnd,
		}
		writer.LogWriter, err = internal.NewFileWriter(l.ctx, &l.waiter, config.logPath, "temp.log", config.maxSize, config.maxAge)
		if err != nil {
			return err
		}
		l.writer.AddWriter(writer)
	case WriteByLevelSeparated:
		for _, level := range allLevels {
			if (l.config.logLevel & level) == level {
				writer := &levelWriter{
					level: level,
				}
				writer.LogWriter, err = internal.NewFileWriter(l.ctx, &l.waiter, pkg.JoinPath(config.logPath, subPath(level)), "temp.log", config.maxSize, config.maxAge)
				if err != nil {
					return err
				}
				l.writer.AddWriter(writer)
			}
		}
	case WriteByBoth:
		targetLevel := make([]Level, 0, 8)
		targetLevel = append(targetLevel, _LevelEnd)
		for _, level := range allLevels {
			if (l.config.logLevel & level) == level {
				targetLevel = append(targetLevel, level)
			}
		}
		for _, level := range targetLevel {
			writer := &levelWriter{
				level: level,
			}
			writer.LogWriter, err = internal.NewFileWriter(l.ctx, &l.waiter, pkg.JoinPath(config.logPath, subPath(level)), "temp.log", config.maxSize, config.maxAge)
			if err != nil {
				return err
			}
			l.writer.AddWriter(writer)
		}
	}
	return nil
}

func (lw *levelWriter) Name() string {
	return subPath(lw.level)
}

func assert(b bool, text string) {
	if b {
		panic(text)
	}
}
