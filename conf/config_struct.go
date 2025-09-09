package conf

import "github.com/spf13/viper"

// DBConfig 是数据库总配置结构体
type DBConfig struct {
	// 数据库类型 mysql mongodb sqlite
	Type string `mapstructure:"db.type"`
	// 日志配置
	*LogConfig
	// MySQL配置
	*MySQLConfig
	// MongoDB配置
	*MongoDBConfig
	// SQLite配置
	*SQLiteConfig
}

type LogConfig struct {
	// 日志文件路径
	Log string `mapstructure:"db.log"`
	// 日志等级
	LogLevel string `mapstructure:"db.loglevel"`
}

func (l *LogConfig) Init() {
	if l == nil {
		return
	}
	if l.Log != "" {
		config.Set("db.log", l.Log)
	}
	if l.LogLevel != "" {
		config.Set("db.loglevel", l.LogLevel)
	}
}

// Init 将总配置设置到config单例上，并调用各数据库配置的Init方法
func (d *DBConfig) Init() {
	if d == nil {
		return
	}

	// 确保config已初始化
	if config == nil {
		config = viper.New()
	}

	// 设置基本配置
	config.Set("db.type", d.Type)

	// 日志初始化
	if d.LogConfig != nil {
		d.LogConfig.Init()
	}
	// 调用各数据库配置的Init方法
	switch d.Type {
	case "mysql":
		if d.MySQLConfig != nil {
			d.MySQLConfig.Init()
		}
	case "sqlite":
		if d.SQLiteConfig != nil {
			d.SQLiteConfig.Init()
		}
	case "mongodb":
		if d.MongoDBConfig != nil {
			d.MongoDBConfig.Init()
		}
	}
}
