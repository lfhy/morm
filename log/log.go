package log

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/lfhy/morm/conf"
	"gorm.io/gorm/logger"
)

var (
	logout  *logger.Interface
	loginit sync.Once
)

func InitDBLoger() logger.Interface {
	if logout != nil {
		return *logout
	}
	LogName := conf.ReadConfigToString("db", "log")
	var logWrite io.Writer
	if LogName == "" {
		logWrite = os.Stdout
	} else {
		os.MkdirAll(filepath.Dir(LogName), 0755)
		f, err := os.OpenFile(LogName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil && err != os.ErrExist {
			log.Println("数据库", "日志初始化失败")
			logWrite = os.Stdout
		} else {
			logWrite = f
		}
	}

	log := logger.New(log.New(logWrite, "\r\n", log.LstdFlags), logger.Config{
		SlowThreshold:             200 * time.Millisecond,                                  // 慢 SQL 阈值
		LogLevel:                  logger.LogLevel(conf.ReadConfigToInt("db", "loglevel")), // 日志级别
		IgnoreRecordNotFoundError: false,                                                   // 忽略ErrRecordNotFound（记录未找到）错误
		Colorful:                  false,                                                   // 禁用彩色打印
	})
	logout = &log
	return *logout
}

func GetDBLoger() logger.Interface {
	if logout == nil {
		loginit.Do(func() {
			InitDBLoger()
		})
	}
	return *logout
}

// 设置自定义日志
func SetDBLoger(log logger.Interface) {
	logout = &log
}
