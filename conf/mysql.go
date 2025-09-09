package conf

import "github.com/spf13/viper"

// MySQLConfig 是MySQL数据库的配置结构体
type MySQLConfig struct {
	// mysql连接数据库
	Database string `mapstructure:"database"`
	// 数据库编码
	Charset string `mapstructure:"charset"`
	// 连接最大生命时间
	ConnMaxLifetime string `mapstructure:"conn_max_lifetime"`
	// mysql连接主机
	Host string `mapstructure:"host"`
	// mysql连接端口
	Port string `mapstructure:"port"`
	// 最大空闲连接数
	MaxIdleConns string `mapstructure:"max_idle_conns"`
	// 最大连接数
	MaxOpenConns string `mapstructure:"max_open_conns"`
	// mysql认证用户
	User string `mapstructure:"user"`
	// mysql认证密码
	Password string `mapstructure:"password"`
}

// Init 将MySQL配置设置到config单例上
func (m *MySQLConfig) Init() {
	if m == nil {
		return
	}

	// 确保config已初始化
	if config == nil {
		config = viper.New()
	}

	config.Set("mysql.database", m.Database)
	config.Set("mysql.charset", m.Charset)
	config.Set("mysql.conn_max_lifetime", m.ConnMaxLifetime)
	config.Set("mysql.host", m.Host)
	config.Set("mysql.port", m.Port)
	config.Set("mysql.max_idle_conns", m.MaxIdleConns)
	config.Set("mysql.max_open_conns", m.MaxOpenConns)
	config.Set("mysql.user", m.User)
	config.Set("mysql.password", m.Password)
}
