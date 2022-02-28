package plogs

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"time"
)

// 全局唯一Logger实例
var defaultLogger *Logger

// Logger 日志记录器
type Logger struct {
	wg           *sync.WaitGroup  // wg
	once         *sync.Once       // once
	closeTag     bool             // 是否关闭
	closeChan    chan struct{}    // 关闭信号量
	flushChan    chan struct{}    // flush buffer chan
	msgChan      chan *logMessage // 接收日志数据的通道
	config       *LogConfig       // 配置
	levelConfigs *levelList       // 需要输出的各个级别对应的配置
}

// NewLogger 初始化defaultLogger
func NewLogger(opts ...Option) (logger *Logger) {
	if defaultLogger == nil {
		defaultLogger = &Logger{
			wg:        &sync.WaitGroup{},
			once:      &sync.Once{},
			closeTag:  false,
			closeChan: make(chan struct{}),
			flushChan: make(chan struct{}),
			config: &LogConfig{
				stdout:        false,
				bufferSize:    10240,
				writeOption:   WriteByMerged,
				cutOption:     CutDaily,
				writeLevel:    0,
				flushDuration: 500 * time.Millisecond,
				appName:       "plogs",
				logPath:       "./logs",
			},
		}
	}
	defaultLogger.once.Do(func() {
		for _, op := range opts {
			op(defaultLogger.config)
		}
		// 初始化
		if err := defaultLogger.init(); err != nil {
			panic(err)
		}
		// 运行
		go defaultLogger.readLoop()
		go defaultLogger.cutLoop()
	})
	return defaultLogger
}

func (log *Logger) Close() {
	if log.closeTag {
		return
	}
	log.closeTag = true
	close(log.msgChan)
	close(log.flushChan)
	close(log.closeChan)
	log.wg.Wait()
	for _, config := range log.levelConfigs.levels {
		config.close()
	}
}

func (log *Logger) init() (err error) {
	// 初始化日志通道
	log.msgChan = make(chan *logMessage, log.config.bufferSize)
	log.levelConfigs = &levelList{
		mu: &sync.Mutex{},
	}
	// 如果需要统一输出到一个日志文件
	logPath := log.config.logPath
	writeLevel := log.config.writeLevel
	writeOption := log.config.writeOption

	if writeOption != WriteByLevel {
		c := &levelConfig{
			level: _LevelEnd,
		}
		if err = c.init(logPath); err != nil {
			return
		}
		log.levelConfigs.levels = append(log.levelConfigs.levels, c)
	}

	// 如果需要根据级别输出到不同的文件
	if writeOption != WriteByMerged {
		// 初始化每个需要输出的日志级别的配置
		allLevels := []Level{LevelFatal, LevelError, LevelWarning, LevelInfo, LevelDebug}
		for _, lv := range allLevels {
			if lv&writeLevel == lv {
				c := &levelConfig{
					level: lv,
				}
				if err = c.init(logPath); err != nil {
					return
				}
				log.levelConfigs.levels = append(log.levelConfigs.levels, c)
			}
		}
	}
	return
}

// 这里开启两个协程，一个负责读取并记录日志，另一个负责切割日志
func (log *Logger) readLoop() {
	log.wg.Add(1)
	ticker := time.NewTimer(log.config.flushDuration)
	defer ticker.Stop()
	for {
		//第一个select，固定周期进行flush，或者周期内通道缓冲达到了容量的90%时进行flush
		//收到flush信号量时需要停止ticker
		//第二个select，收到日志即write，没有日志则进入下一个flush周期
		//每个周期完毕，需要重置ticker
		//日志消息通道关闭时直接return，结束协程
		select {
		case <-ticker.C:
			break
		case _, ok := <-log.flushChan:
			if ok {
				ticker.Stop()
				break
			}
		}
		select {
		case msg, ok := <-log.msgChan:
			log.write(msg)
			if !ok {
				log.wg.Done()
				return
			}
		default:
			break
		}
		ticker.Reset(log.config.flushDuration)
	}
}

func (log *Logger) cutLoop() {
	log.wg.Add(1)
	ticker := time.NewTicker(defaultCutDuration)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			log.cut()
			ticker.Reset(defaultCutDuration)
		case <-log.closeChan:
			log.wg.Done()
			return
		default:
			break
		}
	}
}

// 将消息发送到日志通道
func (log *Logger) log(level Level, message string) {
	if log.closeTag || !level.valid() {
		return
	}
	// 判断是否需要输出
	if (log.config.writeLevel & level) != level {
		return
	}
	// 拼装日志消息, 并发送至chan
	if msgData := log.splicingMessage(level, message); msgData != nil {
		// 如果需要输出到标准输出流，这里同步输出
		if log.config.stdout {
			switch runtime.GOOS {
			case "windows":
				_, _ = fmt.Fprint(os.Stdout, msgData.message)
			default:
				_, _ = fmt.Fprintf(os.Stdout, "%c[%dm%s%c[0m", 0x1B, level.colorCode(), msgData.message, 0x1B)
			}
		}
		log.msgChan <- msgData
	}

	percent := float32(len(log.msgChan)) / float32(log.config.bufferSize)
	if percent > 0.9 {
		log.flushChan <- struct{}{}
	}
}

// 拼接每一条日志
// 每条日志内容格式为:[应用名] [级别前缀] [时间] [文件名:行号] [日志内容]
func (log *Logger) splicingMessage(level Level, message string) (msg *logMessage) {
	_, fileName, line, ok := runtime.Caller(3)
	if !ok {
		return
	}

	// TODO 如何不获取绝对路径，而是项目的相对路径
	//base, _ := filepath.Abs("")
	//fileName = strings.TrimPrefix(fileName, fmt.Sprintf("%s%s", filepath.Dir(base), string(filepath.Separator)))

	//list := filepath.SplitList(fileName)
	//if n := len(list); n > 2 {
	//	fileName = filepath.Join(list[n-2:]...)
	//}

	var (
		appName     = log.config.appName                              // 应用名
		levelPrefix = level.prefix()                                  // 日志级别
		timeDesc    = time.Now().Format("2006/01/02 15:04:05.000000") // 时间
	)

	msg = defaultPool.getLogMessage()
	msg.level = level
	msg.message = fmt.Sprintf("[%s] %s [%s] [%s:%d] %s\n", appName, levelPrefix, timeDesc, fileName, line, message)
	return
}

// 将日志写入句柄
func (log *Logger) write(msgData *logMessage) {
	var msgs []*logMessage

	// 将第一条日志添加进队列中
	if msgData != nil {
		msgs = append(msgs, msgData)
	}
	// 这里将通道中的所有能读取到的日志都读取出来
	select {
	case data, ok := <-log.msgChan:
		if ok {
			msgs = append(msgs, data)
		}
	default:
		break
	}

	mu := log.levelConfigs.mu
	mu.Lock()
	defer mu.Unlock()

	for _, m := range msgs {
		var levels []Level
		var configs []*levelConfig
		switch log.config.writeOption {
		case WriteByLevel:
			levels = append(levels, m.level)
		case WriteByMerged:
			levels = append(levels, _LevelEnd)
		case WriteByAll:
			levels = append(levels, _LevelEnd, m.level)
		}
		configs = log.levelConfigs.getConfig(levels...)
		// 将message写入句柄
		for _, cg := range configs {
			_, _ = fmt.Fprintf(cg, "%s", m.message)
		}
		defaultPool.putMessage(m)
	}
}

func (log *Logger) cut() {
	nowTime := time.Now()
	unix := nowTime.Unix()
	mu := log.levelConfigs.mu

	mu.Lock()
	defer mu.Unlock()

	for _, config := range log.levelConfigs.levels {
		// 判断是否符合切割条件
		switch log.config.cutOption {
		case CutHourly: // 每小时切割一次
			if unix-config.cutTime < 60*60 {
				return
			}
		case CutHalfAnHour: // 半小时切割一次
			if unix-config.cutTime < 30*60 {
				return
			}
		case CutTenMin: // 10 分钟切割一次
			if unix-config.cutTime < 10*60 {
				return
			}
		case CutPer10M: // 每10M切割一次
			if config.size < 5*1024*1024 {
				return
			}
		case CutPer60M: // 每60M切割一次
			if config.size < 60*1024*1024 {
				return
			}
		case CutPer100M: // 每100M切割一次
			if config.size < 100*1024*1024 {
				return
			}
		default: // 默认每天切割
			if unix-config.cutTime < 24*60*60 {
				return
			}
		}
		// 可以切割
		if err := config.reset(); err != nil {
			return
		}

		if log.config.fileMaxTime == 0 && log.config.fileLimit == 0 {
			continue
		}
		// 判断文件数量或者存活时间是否超过限制
		// 文件超过最长保存时间的直接删除
		// 如果剩下的文件数量仍然超过设置的最大保存数量，则删除最旧的文件，保存fileLimit个文件
		existFiles := config.rangeFile(log.config.fileMaxTime, log.config.fileLimit)
		if log.config.fileLimit > 0 && len(existFiles) > log.config.fileLimit {
			sort.Sort(existFiles)
			for i := log.config.fileLimit; i < len(existFiles); i++ {
				_ = os.Remove(filepath.Join(config.filePath, existFiles[i].Name()))
			}
		}
	}
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
