package plogs

import "fmt"

func Panic(args ...interface{}) {
	defaultLogger.panic(args...)
}

func Panicf(template string, args ...interface{}) {
	defaultLogger.panicf(template, args...)
}

func Fatal(args ...interface{}) {
	defaultLogger.fatal(args...)
}

func Fatalf(template string, args ...interface{}) {
	defaultLogger.fatalf(template, args...)
}

func Error(args ...interface{}) {
	defaultLogger.error(args...)
}

func Errorf(template string, args ...interface{}) {
	defaultLogger.errorf(template, args...)
}

func Warn(args ...interface{}) {
	defaultLogger.Warn(args...)
}

func Warnf(template string, args ...interface{}) {
	defaultLogger.Warnf(template, args...)
}

func Info(args ...interface{}) {
	defaultLogger.Info(args...)
}

func Infof(template string, args ...interface{}) {
	defaultLogger.Infof(template, args...)
}

func Debug(args ...interface{}) {
	defaultLogger.Debug(args...)
}

func Debugf(template string, args ...interface{}) {
	defaultLogger.Debugf(template, args...)
}

func getMessage(template string, fmtArgs []interface{}) string {
	if len(fmtArgs) == 0 {
		return template
	}

	if template != "" {
		return fmt.Sprintf(template, fmtArgs...)
	}

	if len(fmtArgs) == 1 {
		if str, ok := fmtArgs[0].(string); ok {
			return str
		}
	}
	return fmt.Sprint(fmtArgs...)
}
