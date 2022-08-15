package main

import (
	"fmt"
	"time"

	"github.com/pyihe/plogs"
)

func main() {
	//opts := []plogs.Option{
	//	plogs.WithName("ALTIMA"),
	//	plogs.WithFileOption(plogs.WriteByLevelMerged),
	//	plogs.WithLogPath("./logs"),
	//	plogs.WithStdout(true),
	//	plogs.WithLogLevel(plogs.LevelInfo | plogs.LevelDebug | plogs.LevelWarn | plogs.LevelError | plogs.LevelFatal | plogs.LevelPanic),
	//	plogs.WithMaxAge(10 * time.Minute),
	//	plogs.WithMaxSize(10 * 1024 * 1024),
	//}
	//
	//logger := plogs.NewLogger(opts...)
	//defer logger.Close()
	//
	//TestLog("hello, I'm %s", "plogs")
	//TestFatal("hello, I'm %s", "plogs")
	//TestDebug("hello, I'm %s", "plogs")

	opts := []plogs.Option{
		plogs.WithName("ALTIMA"),
		plogs.WithFileOption(plogs.WriteByLevelMerged),
		plogs.WithLogPath("./logs"),
		plogs.WithLogLevel(plogs.LevelPanic | plogs.LevelFatal | plogs.LevelError | plogs.LevelWarn | plogs.LevelInfo | plogs.LevelDebug),
		plogs.WithStdout(true),
		plogs.WithMaxAge(24 * time.Hour),
		plogs.WithMaxSize(10 * 1024 * 1024),
	}

	logger := plogs.NewLogger(opts...)
	defer logger.Close()

	plogs.Panic("hello, this is level panic!")
	plogs.Errorf("hello, this is level error")
	plogs.Warnf("hello, this is level warn!")
	plogs.Infof("hello, this is level info!")
	plogs.Debugf("hello, this is level debug!")
	plogs.Fatalf("hello, this is level fatal!")
}

func TestLog(message string, args ...interface{}) {
	tag := time.Now()
	for n := 0; n < 1; n++ {
		for i := 1; i < 500000; i++ {
			time.Sleep(1 * time.Second)
			plogs.Debugf(message, args...)
			plogs.Infof(message, args...)
			plogs.Warnf(message, args...)
			plogs.Errorf(message, args...)
			//plogs.Panicf(message, args...)
			plogs.Fatalf(message, args...)
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

func TestDebug(message string, args ...interface{}) {
	plogs.Debugf(message, args...)
}
