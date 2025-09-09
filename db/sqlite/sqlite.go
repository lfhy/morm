package sqlite

import (
	"time"

	"github.com/glebarez/sqlite"
	"github.com/lfhy/morm/conf"
	"github.com/lfhy/morm/db/sqlorm"
	"github.com/lfhy/morm/log"
	"github.com/lfhy/morm/types"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Configuration struct {
	// 数据库文件路径
	FilePath string
	// 空闲连接数
	MaxIdleConns int
	// 最大连接数
	MaxOpenConns int
	// 连接可复用的时间
	ConnMaxLifetime time.Duration
}

func (c *Configuration) CheckConfig() error {
	if c.FilePath == "" {
		return log.Errorln("SQLite", "数据库文件路径不能为空")
	}

	if c.MaxIdleConns == 0 {
		c.MaxIdleConns = 10
	}

	if c.MaxOpenConns == 0 {
		c.MaxOpenConns = 100
	}

	if c.ConnMaxLifetime == 0 {
		c.ConnMaxLifetime = 30 * time.Minute
	}

	return nil
}

func (c *Configuration) InitDataBase(loger logger.Interface) (*gorm.DB, error) {
	// 连接SQLite
	db, err := gorm.Open(sqlite.Open(c.FilePath), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   loger,
	})
	if err != nil {
		return nil, log.Errorln("SQLite", "数据库连接失败", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, log.Errorln("SQLite", "数据库获取失败", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(c.MaxIdleConns)
	sqlDB.SetMaxOpenConns(c.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(c.ConnMaxLifetime)

	return db, nil
}

func Init(log logger.Interface) (types.ORM, error) {
	c := &Configuration{
		FilePath:        conf.ReadConfigToString("sqlite", "file_path"),
		MaxIdleConns:    conf.ReadConfigToInt("sqlite", "max_idle_conns"),
		MaxOpenConns:    conf.ReadConfigToInt("sqlite", "max_open_conns"),
		ConnMaxLifetime: conf.ReadConfigToTimeDuration("sqlite", "conn_max_lifetime"),
	}

	err := c.CheckConfig()
	if err != nil {
		return nil, err
	}

	conn, err := c.InitDataBase(log)
	if err != nil {
		return nil, err
	}

	sqlorm.ORMConn = &sqlorm.DBConn{DB: conn, AutoMigrate: conf.ReadConfigToBool("db", "auto_create_table")}
	return sqlorm.ORMConn, nil
}
