package db

import (
	"fmt"
	"sync"

	"github.com/lfhy/morm/conf"
	"github.com/lfhy/morm/db/mongodb"
	"github.com/lfhy/morm/db/mysql"
	"gorm.io/gorm/logger"

	"github.com/lfhy/morm/log"

	orm "github.com/lfhy/morm"
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

// 配置日志组件
func SetLogger(l logger.Interface) {
	log.SetDBLoger(l)
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

func Init() (*orm.ORM, error) {
	// 读取配置文件
	err := initConfig()
	if err != nil {
		return nil, err
	}
	// 初始化日志
	log.InitDBLoger()
	var dbconn *orm.ORM
	switch conf.ReadConfigToString("db", "type") {
	case "mysql":
		conn, err := InitMySQL()
		if err != nil {
			return nil, err
		}
		dbconn = &conn
	case "mongodb":
		conn, err := InitMongoDB()
		if err != nil {
			return nil, err
		}
		dbconn = &conn
	}
	return dbconn, nil
}

func InitMongoDB() (orm.ORM, error) {
	// 读取配置文件
	err := initConfig()
	if err != nil {
		return nil, err
	}
	return mongodb.Init()
}

func InitMySQL() (orm.ORM, error) {
	// 读取配置文件
	err := initConfig()
	if err != nil {
		return nil, err
	}
	return mysql.Init(log.GetDBLoger())
}
