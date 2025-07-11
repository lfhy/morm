package morm

import (
	"fmt"
	"sync"

	"github.com/lfhy/morm/conf"
	"github.com/lfhy/morm/db/mongodb"
	"github.com/lfhy/morm/db/mysql"
	"github.com/lfhy/morm/db/sqlite"
	"github.com/lfhy/morm/log"
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

func SetConfig(configFilePath string) {
	configFile = configFilePath
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

func Init(configPath ...string) ORM {
	conn, err := InitWithError(configPath...)
	if err != nil {
		panic(err)
	}
	return conn
}

func InitWithError(configPath ...string) (ORM, error) {
	if len(configPath) > 0 {
		configFile = configPath[0]
	}
	initConfig()
	db := conf.ReadConfigToString("db", "type")
	switch db {
	case "mysql":
		return InitMySQLWithError()
	case "mongodb":
		return InitMongoDBWithError()
	case "sqlite":
		return InitSQLiteWithError()
	}
	return nil, fmt.Errorf("不支持该数据库类型:%v", db)
}

func InitMongoDB(configPath ...string) ORM {
	conn, err := InitMongoDBWithError(configPath...)
	if err != nil {
		panic(err)
	}
	return conn
}

func InitMySQL(configPath ...string) ORM {
	conn, err := InitMySQLWithError(configPath...)
	if err != nil {
		panic(err)
	}
	return conn
}

func InitSQLite(configPath ...string) ORM {
	conn, err := InitSQLiteWithError(configPath...)
	if err != nil {
		panic(err)
	}
	return conn
}

func InitMongoDBWithError(configPath ...string) (ORM, error) {
	if len(configPath) > 0 {
		configFile = configPath[0]
	}
	initConfig()
	return mongodb.Init()
}

func InitMySQLWithError(configPath ...string) (ORM, error) {
	if len(configPath) > 0 {
		configFile = configPath[0]

	}
	initConfig()
	return mysql.Init(log.InitDBLoger())
}

func InitSQLiteWithError(configPath ...string) (ORM, error) {
	if len(configPath) > 0 {
		configFile = configPath[0]
	}
	initConfig()
	return sqlite.Init(log.InitDBLoger())
}
