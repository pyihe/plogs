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
		plogs.WithWriteOption(plogs.WriteByLevel),
		plogs.WithLogPath("./logs"),
		plogs.WithStdout(true),
		plogs.WithWriteLevel(plogs.LevelInfo | plogs.LevelDebug | plogs.LevelWarning | plogs.LevelError | plogs.LevelFatal),
		plogs.WithMaxTime(60 * 60),
		plogs.WithMaxLimit(15),
	}

	logger := plogs.NewLogger(opts...)
	defer logger.Close()

	now := time.Now()
	for i := 1; i <= 100000; i++ {
		//time.Sleep(500 * time.Millisecond)
		plogs.Fatalf("hello, this is output by plogs!")
		plogs.Errorf("hello, this is output by plogs!")
		plogs.Warn("hello, this is output by plogs!")
		plogs.Info("hello, this is output by plogs!")
		plogs.Debugf("hello, this is output by plogs!")
	}
	fmt.Printf("%v\n", time.Now().Sub(now).String())
}
