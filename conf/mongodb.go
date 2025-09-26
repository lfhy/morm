package conf

import "github.com/spf13/viper"

// MongoDBConfig 是MongoDB数据库的配置结构体
type MongoDBConfig struct {
	// 日志配置
	*LogConfig
	// mongodb连接的数据库
	Database string `mapstructure:"mongodb.database"`
	// 连接池大小
	OptionPoolSize string `mapstructure:"mongodb.option_pool_size"`
	// mongodb代理连接方法 必须是socks代理
	Proxy string `mapstructure:"mongodb.proxy"`
	// mongodb连接uri mongodb://[认证用户名]:[认证密码]@[连接地址]/[额外参数]
	Uri string `mapstructure:"mongodb.uri"`
	// w 写关注 默认majority
	W string `mapstructure:"mongodb.w"`
	// ReadMode 读模式 master nearest
	ReadMode string `mapstructure:"mongodb.readmode"`
}

// Init 将MongoDB配置设置到config单例上
func (m *MongoDBConfig) Init() {
	if m == nil {
		return
	}

	// 确保config已初始化
	if config == nil {
		config = viper.New()
	}

	if m.LogConfig != nil {
		m.LogConfig.Init()
	}

	config.Set("mongodb.database", m.Database)
	config.Set("mongodb.option_pool_size", m.OptionPoolSize)
	config.Set("mongodb.proxy", m.Proxy)
	config.Set("mongodb.uri", m.Uri)
	config.Set("mongodb.w", m.W)
	config.Set("mongodb.readmode", m.ReadMode)
}
