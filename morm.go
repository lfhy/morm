package morm

import (
	"fmt"
	glog "log"
	"os"
	"sync"
	"time"

	"github.com/lfhy/morm/conf"
	"github.com/lfhy/morm/db/mongodb"
	"github.com/lfhy/morm/db/mysql"
	"github.com/lfhy/morm/db/sqlite"
	orm "github.com/lfhy/morm/interface"
	"github.com/lfhy/morm/log"
	"gorm.io/gorm/logger"
)

var (
	configInitOnce sync.Once
	configFile     string
)

// 初始化orm配置文件
func InitORMConfig(configFilePath string) (err error) {
	configFile = configFilePath
	return initConfig()
}

func initConfig() (err error) {
	if configFile == "" {
		return fmt.Errorf("配置文件不存在")
	}
	configInitOnce.Do(func() {
		err = conf.InitConfig(configFile)
	})
	return err
}

func Init() orm.ORM {
	var dbconn orm.ORM
	db := conf.ReadConfigToString("db", "type")
	switch db {
	case "mysql":
		conn, err := mysql.Init(InitDBLoger())
		if err != nil {
			panic(err)
		}
		dbconn = conn
	case "mongodb":
		conn, err := mongodb.Init()
		if err != nil {
			panic(err)
		}
		dbconn = conn
	case "sqlite":
		conn, err := sqlite.Init(InitDBLoger())
		if err != nil {
			panic(err)
		}
		dbconn = conn
	}
	return dbconn
}

func InitMongoDB() *ORM {
	conn, err := mongodb.Init()
	if err != nil {
		panic(err)
	}
	return &conn
}

func InitMySQL() *ORM {
	conn, err := mysql.Init(InitDBLoger())
	if err != nil {
		panic(err)
	}
	return &conn
}

func InitSQLite() *ORM {
	conn, err := sqlite.Init(InitDBLoger())
	if err != nil {
		panic(err)
	}
	return &conn
}

func InitDBLoger() logger.Interface {
	LogName := conf.ReadConfigToString("db", "log")
	f, err := os.OpenFile(LogName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil && err != os.ErrExist {
		log.Errorln("数据库", "日志初始化失败")
		f = os.Stdout
	}

	return logger.New(glog.New(f, "\r\n", glog.LstdFlags), logger.Config{
		SlowThreshold:             200 * time.Millisecond,                                  // 慢 SQL 阈值
		LogLevel:                  logger.LogLevel(conf.ReadConfigToInt("db", "loglevel")), // 日志级别
		IgnoreRecordNotFoundError: false,                                                   // 忽略ErrRecordNotFound（记录未找到）错误
		Colorful:                  false,                                                   // 禁用彩色打印
	})
}
