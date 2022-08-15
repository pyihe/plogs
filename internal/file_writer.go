package internal

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/pyihe/go-pkg/syncs"
	"github.com/pyihe/plogs/pkg"
)

const (
	bufferSize = 1024
)

type fileWriter struct {
	ctx         context.Context  // ctx
	wg          *syncs.WgWrapper // waiter
	closed      int32            // writer是否已关闭
	filePath    string           // 文件保存路径
	fileName    string           // 文件名
	maxSize     int64            // 文件大小上限
	currentSize int64            // 当前文件大小（记录当前已经写入的字节数）
	maxAge      time.Duration    // 文件保存最长时间
	file        *os.File         // 文件句柄
	writeBuffer chan []byte      // 写缓存
}

func newFileWriter(ctx context.Context, wg *syncs.WgWrapper, filePath, fileName string, maxSize int64, maxAge time.Duration) *fileWriter {
	if err := pkg.MakeDir(filePath); err != nil {
		panic(err)
	}
	// 打开日志文件句柄
	name := pkg.JoinPathName(filePath, fileName)
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	stat, err := os.Stat(name)
	if err != nil {
		panic(err)
	}
	return &fileWriter{
		closed:      0,
		ctx:         ctx,
		wg:          wg,
		filePath:    filePath,
		fileName:    fileName,
		maxSize:     maxSize,
		maxAge:      maxAge,
		currentSize: stat.Size(),
		file:        file,
		writeBuffer: make(chan []byte, bufferSize),
	}
}

func (fw *fileWriter) write(b []byte) (n int, err error) {
	// 如果已经关闭
	if atomic.LoadInt32(&fw.closed) == 1 {
		return
	}
	fw.writeBuffer <- b
	return len(b), nil
}

func (fw *fileWriter) stop() {
	if atomic.LoadInt32(&fw.closed) == 1 {
		return
	}
	atomic.StoreInt32(&fw.closed, 1)

	fw.clean()
	close(fw.writeBuffer)
}

func (fw *fileWriter) start() {
	fw.wg.Wrap(func() {
		var ticker *time.Ticker
		var duration = fw.checkLife()

		if duration > 0 {
			ticker = time.NewTicker(duration)
		}
		for {
			select {
			case <-fw.ctx.Done(): // 响应最上层调用的Close
				if ticker != nil {
					ticker.Stop()
				}
				return

			case msg := <-fw.writeBuffer: // 写入文件
				fw.writeToFile(msg)
				fw.rotate()
			}
			if ticker != nil {
				select {
				case <-ticker.C: // check目录下的日志文件保存时间，超过maxAge的文件需要删除
					duration = fw.checkLife()
					ticker.Reset(duration)
				default:
					break
				}
			}
		}
	})
}

// 收到Done信号时, 需要将通道内剩余的日志打入文件中
func (fw *fileWriter) clean() {
	count := len(fw.writeBuffer)
	if count == 0 {
		return
	}
	remainMsg := make([][]byte, count)
	index := 0
	for msg := range fw.writeBuffer {
		remainMsg[index] = msg
		index++
		if index == count {
			break
		}
	}
	fw.writeToFile(remainMsg...)

	if fw.maxSize > 0 {
		fw.rotate()
	}
	if fw.maxAge > 0 {
		fw.checkLife()
	}
}

func (fw *fileWriter) writeToFile(msg ...[]byte) {
	for _, m := range msg {
		// 记录到文件中
		n, _ := fw.file.Write(m)
		// 统计当前已经写入文件的字节数
		fw.currentSize += int64(n)
	}
}

func (fw *fileWriter) rotate() {
	if fw.maxSize <= 0 || fw.currentSize < fw.maxSize {
		return
	}
	// 如果写入字节数已经超过最大字节数，则需要切割文件
	// 同步句柄数据到硬盘
	fw.file.Sync()

	// 关闭句柄
	fw.file.Close()

	// 重命名
	nowTime := time.Now()
	oldName := fw.file.Name()
	newName := pkg.JoinPathName(fw.filePath, fmt.Sprintf("%s.log", nowTime.Format("2006_01_02_15_04_05")))
	os.Rename(oldName, newName)

	// 重置size和句柄
	fw.currentSize = 0
	fw.file, _ = os.OpenFile(oldName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
}

func (fw *fileWriter) checkLife() time.Duration {
	// 如果没有设置文件最大保存时长，则不再校验
	if fw.maxAge <= 0 {
		return 0
	}

	const duration = 30 * time.Second

	filepath.Walk(fw.filePath, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		// 排除后缀不为.log的文件以及正在pending的文件
		if filepath.Ext(path) != ".log" || info.Name() == fw.fileName {
			return nil
		}
		// 排除没有达到maxSize的文件
		if info.Size() < fw.maxSize {
			return nil
		}
		if time.Now().Sub(info.ModTime()) < fw.maxAge {
			return nil
		}
		return os.Remove(path)
	})
	return duration
}
