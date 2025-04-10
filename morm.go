package morm

import (
	"fmt"
	"sync"

	"github.com/lfhy/morm/conf"
	"github.com/lfhy/morm/db/mongodb"
	"github.com/lfhy/morm/db/mysql"
	"github.com/lfhy/morm/db/sqlite"
	orm "github.com/lfhy/morm/interface"
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
		conn, err := mysql.Init(log.InitDBLoger())
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
		conn, err := sqlite.Init(log.InitDBLoger())
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
	conn, err := mysql.Init(log.InitDBLoger())
	if err != nil {
		panic(err)
	}
	return &conn
}

func InitSQLite() *ORM {
	conn, err := sqlite.Init(log.InitDBLoger())
	if err != nil {
		panic(err)
	}
	return &conn
}
