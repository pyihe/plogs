package plogs

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/pyihe/go-pkg/files"
)

// 所有可输出级别对应的配置
type levelList struct {
	mu     *sync.Mutex
	levels []*levelConfig
}

func (ll *levelList) getConfig(levels ...Level) (configs []*levelConfig) {
	for _, l := range levels {
		for _, c := range ll.levels {
			if c.level == l {
				configs = append(configs, c)
			}
		}
	}
	return
}

// 每个Level对应的配置
type levelConfig struct {
	level    Level    // 日志级别
	prefix   string   // 日志前缀
	filePath string   // 日志文件存放路径
	fileName string   // 文件名
	cutTime  int64    // 文件切割时间
	size     int64    // 写入的大小
	fd       *os.File // 文件句柄
}

func (lc *levelConfig) init(root string) (err error) {
	nowTime := time.Now()
	lc.prefix = lc.level.prefix()
	lc.filePath = filepath.Join(root, lc.level.subPath())
	lc.cutTime = nowTime.Unix()
	lc.size = 0
	lc.fileName = "temp.log"

	// 创建目录(如果不存在的话)
	if err = files.NewPath(lc.filePath); err != nil {
		return
	}
	// 打开文件
	lc.fd, err = os.OpenFile(filepath.Join(lc.filePath, lc.fileName), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
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
	lc.size = 0

	// 5. 重置fd
	lc.fd, err = os.OpenFile(oldPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	return
}

func (lc *levelConfig) close() {
	_ = lc.fd.Sync()
	_ = lc.fd.Close()
}
