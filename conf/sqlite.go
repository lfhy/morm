package conf

import "github.com/spf13/viper"

// SQLiteConfig 是SQLite数据库的配置结构体
type SQLiteConfig struct {
	// 数据库路径
	// 这里传入的是gorm的DSN 支持内存模式等其他特性 如:"file:testdatabase?mode=memory&cache=shared"
	// 也可以传入数据库的存储路径
	FilePath string `mapstructure:"file_path"`
	// 最大空闲连接数
	MaxIdleConns string `mapstructure:"max_idle_conns"`
	// 最大连接数
	MaxOpenConns string `mapstructure:"max_open_conns"`
	// 连接最大生命时间
	ConnMaxLifetime string `mapstructure:"conn_max_lifetime"`
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
	
	config.Set("sqlite.file_path", s.FilePath)
	config.Set("sqlite.max_idle_conns", s.MaxIdleConns)
	config.Set("sqlite.max_open_conns", s.MaxOpenConns)
	config.Set("sqlite.conn_max_lifetime", s.ConnMaxLifetime)
}