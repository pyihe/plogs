package plogs

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/pyihe/go-pkg/files"
)

type (
	// 描述每一条日志信息，包括级别和内容
	logMessage struct {
		level   Level  // 日志级别
		message string // 日志内容
	}

	// 每个Level对应的配置
	levelConfig struct {
		level    Level    // 日志级别
		prefix   string   // 日志前缀
		filePath string   // 日志文件存放路径
		fileName string   // 文件名
		cutTime  int64    // 文件切割时间
		fd       *os.File // 文件句柄
	}
)

func (lc *levelConfig) init(root string) (err error) {
	lc.prefix = lc.level.prefix()
	nowTime := time.Now()
	lc.filePath = filepath.Join(root, lc.level.subPath())
	lc.cutTime = nowTime.Unix()
	lc.fileName = "temp.log"

	// 创建目录(如果不存在的话)
	if err = files.NewPath(lc.filePath); err != nil {
		return
	}
	// 打开文件
	lc.fd, err = os.OpenFile(filepath.Join(lc.filePath, lc.fileName), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	return
}

func (lc *levelConfig) reset() (err error) {
	// 1. 将数据flush进硬盘
	if err = lc.fd.Sync(); err != nil {
		return
	}
	// 2. 关闭fd并清空fd
	_ = lc.fd.Close()
	lc.fd = nil

	// 3. 将文件重命名
	nowTime := time.Now()
	oldPath := filepath.Join(lc.filePath, lc.fileName)
	newPath := filepath.Join(lc.filePath, fmt.Sprintf("%s.log", nowTime.Format("2006_01_02_15_04_05")))
	if err = os.Rename(oldPath, newPath); err != nil {
		return
	}
	// 4. 更新切割时间
	lc.cutTime = nowTime.Unix()

	// 5. 重置fd
	lc.fd, err = os.OpenFile(filepath.Join(lc.filePath, lc.fileName), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	return
}

func (lc *levelConfig) close() {
	_ = lc.fd.Sync()
	_ = lc.fd.Close()
}

var (
	defaultLogger *Logger
)

// Logger 日志记录器
type Logger struct {
	wg                *sync.WaitGroup        // wg
	once              *sync.Once             // once
	stdTag            bool                   // 是否在终端输出日志信息
	closed            bool                   // 是否关闭
	msgChanBufferSize int                    // 日志数据通道缓存长度
	writeOption       WriteOption            // 日志文件记录方式
	cutOption         CutOption              // 日志文件切割周期
	writeLevel        Level                  // 需要记录的日志级别
	flushDuration     time.Duration          // 日志通道缓冲区flush周期，单位毫秒
	appName           string                 // 应用名称
	logSavePath       string                 // 日志存放目录: /var/log
	flushChan         chan struct{}          // 日志通道缓冲flush信号量通道，收到信号量时立即从msgChan中将日志读走
	msgChan           chan *logMessage       // 接收日志数据的通道
	levelConfigs      map[Level]*levelConfig // 每个级别对应的配置
}

// NewLogger 初始化defaultLogger
func NewLogger(opts ...Option) (logger *Logger) {
	if defaultLogger == nil {
		defaultLogger = &Logger{
			wg:                &sync.WaitGroup{},
			once:              &sync.Once{},
			stdTag:            true,
			msgChanBufferSize: 1024,
			flushDuration:     1000,
			writeOption:       WriteByMerged,
			cutOption:         CutDaily,
			writeLevel:        0,
			logSavePath:       "logs",
			flushChan:         make(chan struct{}),
			msgChan:           make(chan *logMessage, 1024),
			levelConfigs:      make(map[Level]*levelConfig),
		}
	}
	defaultLogger.once.Do(func() {
		for _, op := range opts {
			op(defaultLogger)
		}
		// 初始化
		if err := defaultLogger.init(); err != nil {
			panic(err)
		}
		// 运行
		defaultLogger.run()
	})
	return defaultLogger
}

func (log *Logger) Close() {
	log.closed = true
	close(log.msgChan)
	close(log.flushChan)
	log.wg.Wait()
	for _, config := range log.levelConfigs {
		config.close()
	}
}

func (log *Logger) init() (err error) {
	// 如果需要统一输出到一个日志文件
	if log.writeOption != WriteByLevel {
		c := &levelConfig{
			level: _LevelEnd,
		}
		if err = c.init(log.logSavePath); err != nil {
			return
		}
		log.levelConfigs[_LevelEnd] = c
	}

	// 如果需要根据级别输出到不同的文件
	if log.writeOption != WriteByMerged {
		// 初始化每个需要输出的日志级别的配置
		allLevels := []Level{LevelFatal, LevelError, LevelWarning, LevelInfo, LevelDebug}
		for _, lv := range allLevels {
			if (lv & log.writeLevel) != lv {
				continue
			}
			c := &levelConfig{
				level: lv,
			}
			if err = c.init(log.logSavePath); err != nil {
				return
			}
			log.levelConfigs[lv] = c
		}
	}
	return
}

func (log *Logger) run() {
	log.wg.Add(1)
	go func(wg *sync.WaitGroup) {
		ticker := time.NewTimer(log.flushDuration)
		defer ticker.Stop()
		for {
			/*
				第一个select，固定周期进行flush，或者周期内通道缓冲达到了容量的90%时进行flush
				收到flush信号量时需要停止ticker
				第二个select，收到日志即write，没有日志则进入下一个flush周期
				每个周期完毕，需要重置ticker
				日志消息通道关闭时直接return，结束协程
			*/
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
					wg.Done()
					return
				}
			default:
				break
			}
			ticker.Reset(log.flushDuration)
		}
	}(log.wg)
}

// 将消息发送到日志通道
func (log *Logger) syncMessage(level Level, message string) {
	if log.closed || !level.valid() {
		return
	}
	// 判断是否需要输出
	if (log.writeLevel & level) != level {
		return
	}
	// 拼装日志消息, 并发送至chan
	if msgData := log.splicingMessage(level, message); msgData != nil {
		// 如果需要输出到标准输出流，这里同步输出
		if log.stdTag {
			_, _ = fmt.Fprintf(os.Stdout, "%c[%dm%s%c[0m\n", 0x1B, msgData.level.colorCode(), msgData.message, 0x1B)
		}
		log.msgChan <- msgData
	}

	percent := float32(len(log.msgChan)) / float32(log.msgChanBufferSize)
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

	list := filepath.SplitList(fileName)
	if n := len(list); n > 2 {
		fileName = filepath.Join(list[n-2:]...)
	}

	var (
		appName     = log.appName                                     // 应用名
		levelPrefix = level.prefix()                                  // 日志级别
		timeDesc    = time.Now().Format("2006/01/02 15:04:05.000000") // 时间
	)

	if appName == "" {
		appName = "plogs"
	}

	msg = &logMessage{
		level:   level,
		message: fmt.Sprintf("[%s] %s [%s] [%s:%d] %s", appName, levelPrefix, timeDesc, fileName, line, message),
	}
	return
}

// 将日志写入句柄
func (log *Logger) write(msgData *logMessage) {
	var targetLevels = make(map[Level]*levelConfig)
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
	for _, m := range msgs {
		var configs []*levelConfig
		var cs = log.levelConfigs

		switch log.writeOption {
		case WriteByLevel:
			configs = append(configs, cs[m.level])
		case WriteByMerged:
			configs = append(configs, cs[_LevelEnd])
		case WriteByAll:
			configs = append(configs, cs[_LevelEnd], cs[m.level])
		}
		// 将message写入句柄
		for _, cg := range configs {
			_, _ = cg.fd.WriteString(m.message)
			_, _ = cg.fd.WriteString("\n")
		}
	}

	// 判断是否需要切割文件了
	log.cut(targetLevels)
}

func (log *Logger) cut(configs map[Level]*levelConfig) {
	nowTime := time.Now()
	unix := nowTime.Unix()

	for _, config := range configs {
		fileInfo, err := config.fd.Stat()
		if err != nil {
			return
		}
		size := fileInfo.Size()

		// 判断是否符合切割条件
		switch log.cutOption {
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
			if size < 10*1024*1024 {
				return
			}
		case CutPer60M: // 每60M切割一次
			if size < 60*1024*1024 {
				return
			}
		case CutPer100M: // 每100M切割一次
			if size < 100*1024*1024 {
				return
			}
		default: // 默认每天切割
			if unix-config.cutTime < 24*60*60 {
				return
			}
		}
		// 可以切割
		if err = config.reset(); err != nil {
			return
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
