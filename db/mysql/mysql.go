package mysql

import (
	"fmt"
	"time"

	"github.com/lfhy/morm/log"

	"github.com/lfhy/morm/conf"
	"github.com/lfhy/morm/db/sqlorm"

	"github.com/lfhy/morm/types"

	gmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Configuration struct {
	// 连接主机
	Host string
	// 连接端口
	Port string
	// 连接用户
	UserName string
	// 连接密码
	Password string
	// 数据库名称
	DataBase string
	// 数据库编码格式
	Charset string
	// 空闲连接数
	MaxIdleConns int
	// 最大连接数
	MaxOpenConns int
	// 连接可复用的时间
	ConnMaxLifetime time.Duration
}

func (c *Configuration) CheckConfig() error {
	if c.DataBase == "" {
		return log.Errorln("MYSQL", "数据库不能为空")
	}

	if c.Host == "" {
		c.Host = "127.0.0.1"
	}

	if c.Port == "" {
		c.Port = "3306"
	}

	if c.UserName == "" {
		c.UserName = "root"
	}

	if c.Charset == "" {
		c.Charset = "utf8mb4"
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
	// 生成连接字符串
	dsn := fmt.Sprintf("tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local", c.Host, c.Port, c.DataBase, c.Charset)
	if c.Password != "" {
		dsn = fmt.Sprintf("%s:%s@%s", c.UserName, c.Password, dsn)
	} else {
		dsn = fmt.Sprintf("%s@%s", c.UserName, dsn)
	}

	// 生成mysql配置
	mysqlConfig := gmysql.Config{
		DSN:                       dsn,   // 连接字符串
		DefaultStringSize:         256,   // string 类型字段的默认长度
		DisableDatetimePrecision:  true,  // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,  // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,  // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false, // 根据版本自动配置
	}

	// 连接mysql
	db, err := gorm.Open(gmysql.New(mysqlConfig), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,  // 禁用自动创建外键约束
		Logger:                                   loger, // 使用自定义 Logger
	})
	if err != nil {
		return nil, log.Errorln("MySQL", "数据库连接失败", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, log.Errorln("MySQL", "数据库获取失败", err)
	}
	// 设置优化参数
	sqlDB.SetMaxIdleConns(c.MaxIdleConns)
	sqlDB.SetMaxOpenConns(c.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(c.ConnMaxLifetime)
	return db, nil
}

func Init(log logger.Interface) (types.ORM, error) {
	c := &Configuration{
		Host:            conf.ReadConfigToString("mysql", "host"),
		Port:            conf.ReadConfigToString("mysql", "port"),
		UserName:        conf.ReadConfigToString("mysql", "user"),
		Password:        conf.ReadConfigToString("mysql", "password"),
		DataBase:        conf.ReadConfigToString("mysql", "database"),
		Charset:         conf.ReadConfigToString("mysql", "charset"),
		MaxIdleConns:    conf.ReadConfigToInt("mysql", "max_idle_conns"),
		MaxOpenConns:    conf.ReadConfigToInt("mysql", "max_open_conns"),
		ConnMaxLifetime: conf.ReadConfigToTimeDuration("mysql", "conn_max_lifetime"),
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
	// 校验数据库
	// sqlorm.ORMConn.CheckDB()
	return sqlorm.ORMConn, nil
}
