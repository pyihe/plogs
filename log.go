package plogs

func (log *Logger) Panic(args ...interface{}) {
	message := getMessage("", args)
	log.log(LevelPanic, message)
	log.panic(message)
}

func (log *Logger) Panicf(template string, args ...interface{}) {
	message := getMessage(template, args)
	log.log(LevelPanic, message)
	log.panic(message)
}

func (log *Logger) Fatal(args ...interface{}) {
	message := getMessage("", args)
	log.log(LevelFatal, message)
	log.exit()
}

func (log *Logger) Fatalf(template string, args ...interface{}) {
	message := getMessage(template, args)
	log.log(LevelFatal, message)
	log.exit()
}

func (log *Logger) Error(args ...interface{}) {
	message := getMessage("", args)
	log.log(LevelError, message)
}

func (log *Logger) Errorf(template string, args ...interface{}) {
	message := getMessage(template, args)
	log.log(LevelError, message)
}

func (log *Logger) Warn(args ...interface{}) {
	message := getMessage("", args)
	log.log(LevelWarning, message)
}

func (log *Logger) Warnf(template string, args ...interface{}) {
	message := getMessage(template, args)
	log.log(LevelWarning, message)
}

func (log *Logger) Info(args ...interface{}) {
	message := getMessage("", args)
	log.log(LevelInfo, message)
}

func (log *Logger) Infof(template string, args ...interface{}) {
	message := getMessage(template, args)
	defaultLogger.log(LevelInfo, message)
}

func (log *Logger) Debug(args ...interface{}) {
	message := getMessage("", args)
	log.log(LevelDebug, message)
}

func (log *Logger) Debugf(template string, args ...interface{}) {
	message := getMessage(template, args)
	log.log(LevelDebug, message)
}

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
