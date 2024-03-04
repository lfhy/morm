package morm

import (
	"fmt"
	"sync"

	"github.com/lfhy/morm/conf"
	"github.com/lfhy/morm/db/mongodb"
	"github.com/lfhy/morm/db/mysql"
	"gorm.io/gorm/logger"

	"github.com/lfhy/morm/log"

	orm "github.com/lfhy/morm/interface"
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

func Init() (orm.ORM, error) {
	// 读取配置文件
	err := initConfig()
	if err != nil {
		return nil, err
	}
	// 初始化日志
	log.InitDBLoger()
	switch conf.ReadConfigToString("db", "type") {
	case "mysql":
		return InitMySQL()

	case "mongodb":
		return InitMongoDB()
	}
	return nil, fmt.Errorf("不支持的数据库类型:%v", conf.ReadConfigToString("db", "type"))
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
