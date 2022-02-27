package plogs

import "sync"

// 描述每一条日志信息，包括级别和内容
type logMessage struct {
	level   Level  // 日志级别
	message string // 日志内容
}

// logMessage池子
type messagePool struct {
	pool sync.Pool
}

var defaultPool messagePool

func (p *messagePool) getLogMessage() *logMessage {
	data := p.pool.Get()
	if data != nil {
		return data.(*logMessage)
	}
	return &logMessage{
		level:   _LevelBegin,
		message: "",
	}
}

func (p *messagePool) putMessage(msg *logMessage) {
	if msg != nil {
		msg.level = _LevelBegin
		msg.message = ""
		p.pool.Put(msg)
	}

}
