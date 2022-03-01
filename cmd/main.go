package main

import (
	"fmt"
	"time"

	"github.com/pyihe/plogs"
)

func main() {
	opts := []plogs.Option{
		plogs.WithAppName("ALTIMA"),
		plogs.WithBufferSize(10240),
		plogs.WithCutOption(plogs.CutPer10M),
		plogs.WithFlushDuration(500 * time.Millisecond),
		plogs.WithWriteOption(plogs.WriteByAll),
		plogs.WithLogPath("./logs"),
		plogs.WithStdout(true),
		plogs.WithWriteLevel(plogs.LevelInfo | plogs.LevelDebug | plogs.LevelWarning | plogs.LevelError | plogs.LevelFatal | plogs.LevelPanic),
		plogs.WithMaxTime(24 * 60 * 60),
	}

	logger := plogs.NewLogger(opts...)
	defer logger.Close()

	TestLog("hello, I'm %s", "plogs")
	//TestFatal("hello, I'm %s", "plogs")
	//TestPanic("hello, I'm %s", "plogs")
}

func TestLog(message string, args ...interface{}) {
	tag := time.Now()
	for n := 0; n < 3; n++ {
		for i := 1; i < 500000; i++ {
			plogs.Debugf(message, args...)
			plogs.Infof(message, args...)
			plogs.Warnf(message, args...)
			plogs.Errorf(message, args...)
		}
	}
	fmt.Printf("time consume: %v\n", time.Now().Sub(tag).Milliseconds())
}

func TestFatal(message string, args ...interface{}) {
	plogs.Fatalf(message, args...)
}

func TestPanic(message string, args ...interface{}) {
	plogs.Panicf(message, args...)
}
