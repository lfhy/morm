package morm

import (
	"fmt"
	"sync"

	"github.com/lfhy/morm/conf"
	"github.com/lfhy/morm/db/mongodb"
	"github.com/lfhy/morm/db/mysql"
	"github.com/lfhy/morm/db/sqlite"
	"github.com/lfhy/morm/log"
	"github.com/spf13/viper"
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

// 设置viper配置
func UseViperConfig(config *viper.Viper) {
	conf.SetViperConfig(config)
}

// 获取当前viper配置实例
func GetConfig() *viper.Viper {
	return conf.GetConfig()
}

// 初始化配置(单次执行)
func initConfig() (err error) {
	if conf.GetConfig() != nil {
		return nil
	}
	if configFile == "" {
		return fmt.Errorf("配置文件不存在")
	}
	configInitOnce.Do(func() {
		err = conf.InitConfig(configFile)
	})
	return err
}

// 初始化ORM连接(带错误panic)
func Init(configPath ...string) ORM {
	conn, err := InitWithError(configPath...)
	if err != nil {
		panic(err)
	}
	return conn
}

// 初始化ORM连接(返回错误)
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

// 初始化MongoDB连接(带错误panic)
func InitMongoDB(configPath ...string) ORM {
	conn, err := InitMongoDBWithError(configPath...)
	if err != nil {
		panic(err)
	}
	return conn
}

// 初始化MySQL连接(带错误panic)
func InitMySQL(configPath ...string) ORM {
	conn, err := InitMySQLWithError(configPath...)
	if err != nil {
		panic(err)
	}
	return conn
}

// 初始化SQLite连接(带错误panic)
func InitSQLite(configPath ...string) ORM {
	conn, err := InitSQLiteWithError(configPath...)
	if err != nil {
		panic(err)
	}
	return conn
}

// 初始化MongoDB连接(返回错误)
func InitMongoDBWithError(configPath ...string) (ORM, error) {
	if len(configPath) > 0 {
		configFile = configPath[0]
	}
	initConfig()
	return mongodb.Init()
}

// 初始化MySQL连接(返回错误)
func InitMySQLWithError(configPath ...string) (ORM, error) {
	if len(configPath) > 0 {
		configFile = configPath[0]
	}
	initConfig()
	return mysql.Init(log.InitDBLoger())
}

// 初始化SQLite连接(返回错误)
func InitSQLiteWithError(configPath ...string) (ORM, error) {
	if len(configPath) > 0 {
		configFile = configPath[0]
	}
	initConfig()
	return sqlite.Init(log.InitDBLoger())
}

// 使用配置结构体初始化
func InitWithDBConfig(config *DBConfig) ORM {
	config.Init()
	return Init(configFile)
}

// 使用配置结构体初始化
func InitWithDBConfigWithError(config *DBConfig) (ORM, error) {
	config.Init()
	return InitWithError(configFile)
}

// 使用配置结构体初始化MySQL
func InitMySQLWithDBConfig(config *MySQLConfig) ORM {
	config.Init()
	return InitMySQL()
}

// 使用配置结构体初始化MySQL
func InitMySQLWithDBConfigWithError(config *MySQLConfig) (ORM, error) {
	config.Init()
	return InitMySQLWithError()
}

// 使用配置结构体初始化SQLite
func InitSQLiteWithDBConfig(config *SQLiteConfig) ORM {
	config.Init()
	return InitSQLite()
}

// 使用配置结构体初始化SQLite
func InitSQLiteWithDBConfigWithError(config *SQLiteConfig) (ORM, error) {
	config.Init()
	return InitSQLiteWithError()
}

func InitMongoDBWithDBConfig(config *MongoDBConfig) ORM {
	config.Init()
	return InitMongoDB()
}

func InitMongoDBWithDBConfigWithError(config *MongoDBConfig) (ORM, error) {
	config.Init()
	return InitMongoDBWithError()
}
