package conf

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

var config *viper.Viper

// 初始化配置文件
func InitConfig(conf string) error {
	config = viper.New()
	config.SetConfigType("toml")
	config.SetConfigFile(conf)
	return config.ReadInConfig()
}

// 读取配置文件中的string值
func ReadConfigToString(title, key string) string {
	return config.GetString(fmt.Sprintf("%v.%v", title, key))
}

// 读取配置文件中的int值
func ReadConfigToInt(title, key string) int {
	return config.GetInt(fmt.Sprintf("%v.%v", title, key))
}

// 读取配置文件中的时间值
func ReadConfigToTimeDuration(title, key string) time.Duration {
	return config.GetDuration(fmt.Sprintf("%v.%v", title, key))
}
