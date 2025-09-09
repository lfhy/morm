package conf

import "github.com/spf13/viper"

// DBConfig 是数据库总配置结构体
type DBConfig struct {
	// 数据库类型 mysql mongodb sqlite
	Type string `mapstructure:"type"`
	// 日志文件路径
	Log string `mapstructure:"log"`
	// 日志等级
	LogLevel string `mapstructure:"loglevel"`
	// MySQL配置
	MySQLConfig *MySQLConfig `mapstructure:"mysql"`
	// MongoDB配置
	MongoDBConfig *MongoDBConfig `mapstructure:"mongodb"`
	// SQLite配置
	SQLiteConfig *SQLiteConfig `mapstructure:"sqlite"`
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
	config.Set("type", d.Type)
	config.Set("log", d.Log)
	config.Set("loglevel", d.LogLevel)

	// 调用各数据库配置的Init方法
	if d.MySQLConfig != nil {
		d.MySQLConfig.Init()
	}

	if d.MongoDBConfig != nil {
		d.MongoDBConfig.Init()
	}

	if d.SQLiteConfig != nil {
		d.SQLiteConfig.Init()
	}
}
