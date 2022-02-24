package plogs

func Fatal(args ...interface{}) {
	message := getMessage("", args)
	defaultLogger.syncMessage(LevelFatal, message)
}

func Fatalf(template string, args ...interface{}) {
	message := getMessage(template, args)
	defaultLogger.syncMessage(LevelFatal, message)
}

func Error(args ...interface{}) {
	message := getMessage("", args)
	defaultLogger.syncMessage(LevelError, message)
}

func Errorf(template string, args ...interface{}) {
	message := getMessage(template, args)
	defaultLogger.syncMessage(LevelError, message)
}

func Warn(args ...interface{}) {
	message := getMessage("", args)
	defaultLogger.syncMessage(LevelWarning, message)
}

func Warnf(template string, args ...interface{}) {
	message := getMessage(template, args)
	defaultLogger.syncMessage(LevelWarning, message)
}

func Info(args ...interface{}) {
	message := getMessage("", args)
	defaultLogger.syncMessage(LevelInfo, message)
}

func Infof(template string, args ...interface{}) {
	message := getMessage(template, args)
	defaultLogger.syncMessage(LevelInfo, message)
}

func Debug(args ...interface{}) {
	message := getMessage("", args)
	defaultLogger.syncMessage(LevelDebug, message)
}

func Debugf(template string, args ...interface{}) {
	message := getMessage(template, args)
	defaultLogger.syncMessage(LevelDebug, message)
}
