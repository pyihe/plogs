package plogs

func Fatal(args ...interface{}) {
	message := getMessage("", args)
	defaultLogger.log(LevelFatal, message)
}

func Fatalf(template string, args ...interface{}) {
	message := getMessage(template, args)
	defaultLogger.log(LevelFatal, message)
}

func Error(args ...interface{}) {
	message := getMessage("", args)
	defaultLogger.log(LevelError, message)
}

func Errorf(template string, args ...interface{}) {
	message := getMessage(template, args)
	defaultLogger.log(LevelError, message)
}

func Warn(args ...interface{}) {
	message := getMessage("", args)
	defaultLogger.log(LevelWarning, message)
}

func Warnf(template string, args ...interface{}) {
	message := getMessage(template, args)
	defaultLogger.log(LevelWarning, message)
}

func Info(args ...interface{}) {
	message := getMessage("", args)
	defaultLogger.log(LevelInfo, message)
}

func Infof(template string, args ...interface{}) {
	message := getMessage(template, args)
	defaultLogger.log(LevelInfo, message)
}

func Debug(args ...interface{}) {
	message := getMessage("", args)
	defaultLogger.log(LevelDebug, message)
}

func Debugf(template string, args ...interface{}) {
	message := getMessage(template, args)
	defaultLogger.log(LevelDebug, message)
}
