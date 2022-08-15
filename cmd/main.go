package main

import (
	"fmt"
	"time"

	"github.com/pyihe/plogs"
)

func main() {
	opts := []plogs.Option{
		plogs.WithName("ALTIMA"),
		plogs.WithFileOption(plogs.WriteByLevelMerged),
		plogs.WithLogPath("./logs"),
		plogs.WithStdout(true),
		plogs.WithLogLevel(plogs.LevelInfo | plogs.LevelDebug | plogs.LevelWarn | plogs.LevelError | plogs.LevelFatal | plogs.LevelPanic),
		plogs.WithMaxAge(24 * time.Hour),
		plogs.WithMaxSize(60 * 1024 * 1024),
	}

	logger := plogs.NewLogger(opts...)
	//defer logger.Close()

	go func() {
		TestLog("hello, I'm %s", "plogs")
	}()
	time.Sleep(3 * time.Second)
	logger.Close()
}

func TestLog(message string, args ...interface{}) {
	tag := time.Now()
	for n := 0; n < 3; n++ {
		for i := 1; i < 500000; i++ {
			go plogs.Debugf(message, args...)
			go plogs.Infof(message, args...)
			go plogs.Warnf(message, args...)
			go plogs.Errorf(message, args...)
			//plogs.Panicf(message, args...)
			//plogs.Fatalf(message, args...)
		}
	}
	fmt.Printf("time consume: %v\n", time.Now().Sub(tag).Milliseconds())
}

func TestFatal(message string, args ...interface{}) {
	plogs.Fatalf(message, args...)
}
