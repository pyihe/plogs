package plogs

func Panic(args ...interface{}) {
	message := getMessage("", args)
	defaultLogger.log(LevelPanic, message)
	defaultLogger.panic(message)
}

func Panicf(template string, args ...interface{}) {
	message := getMessage(template, args)
	defaultLogger.log(LevelPanic, message)
	defaultLogger.panic(message)
}

func Fatal(args ...interface{}) {
	message := getMessage("", args)
	defaultLogger.log(LevelFatal, message)
	defaultLogger.exit()
}

func Fatalf(template string, args ...interface{}) {
	message := getMessage(template, args)
	defaultLogger.log(LevelFatal, message)
	defaultLogger.exit()
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
