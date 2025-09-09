package conf

import "github.com/spf13/viper"

// SQLiteConfig 是SQLite数据库的配置结构体
type SQLiteConfig struct {
	// 日志配置
	*LogConfig
	// 自动创表
	AutoCreateTable bool `mapstructure:"db.auto_create_table"`
	// 数据库路径
	// 这里传入的是gorm的DSN 支持内存模式等其他特性 如:"file:testdatabase?mode=memory&cache=shared"
	// 也可以传入数据库的存储路径
	FilePath string `mapstructure:"sqlite.file_path"`
	// 最大空闲连接数
	MaxIdleConns string `mapstructure:"sqlite.max_idle_conns"`
	// 最大连接数
	MaxOpenConns string `mapstructure:"sqlite.max_open_conns"`
	// 连接最大生命时间
	ConnMaxLifetime string `mapstructure:"sqlite.conn_max_lifetime"`
}

// Init 将SQLite配置设置到config单例上
func (s *SQLiteConfig) Init() {
	if s == nil {
		return
	}

	// 确保config已初始化
	if config == nil {
		config = viper.New()
	}

	if s.LogConfig != nil {
		s.LogConfig.Init()
	}
	config.Set("db.auto_create_table", s.AutoCreateTable)

	config.Set("sqlite.file_path", s.FilePath)
	config.Set("sqlite.max_idle_conns", s.MaxIdleConns)
	config.Set("sqlite.max_open_conns", s.MaxOpenConns)
	config.Set("sqlite.conn_max_lifetime", s.ConnMaxLifetime)
}
