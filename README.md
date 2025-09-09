# Morm
混合ORM(Mixed orm)，配置数据源源即可轻松访问不同类型的数据库，目前支持```mysql```、```mongodb```和```sqlite```，底层使用```gorm```和```mongo-driver```实现。

配置文件使用```viper```进行解析

# 通过配置结构体初始化

除了使用配置文件初始化ORM外，还可以通过DBConfig结构体直接初始化ORM:

```golang
package main

import (
	"github.com/lfhy/morm"
	"github.com/lfhy/morm/conf"
)

func main() {
	// 通过结构体配置初始化
	dbConfig := &conf.DBConfig{
		Type: "sqlite",
		LogConfig: &conf.LogConfig{
			Log:      "./db.log",
			LogLevel: "4",
		},
		SQLiteConfig: &conf.SQLiteConfig{
			FilePath:        "./test.db",
			ConnMaxLifetime: "1h",
			MaxIdleConns:    "10",
			MaxOpenConns:    "100",
		},
	}

	// 使用配置结构体初始化ORM
	orm := morm.InitWithDBConfig(dbConfig)
	
	// 后续使用ORM进行数据库操作...
}
```

# 配置文件参考
```toml
[db]
log = './db.log'    # 日志文件路径
loglevel = '4'  # 日志等级 
type = 'mysql' # 默认orm类型

[mongodb]
# mongodb连接的数据库
database = 'testorm'    
# 连接池大小
option_pool_size = '200'   
# mongodb代理连接方法
proxy = 'socks5://127.0.0.1:7890'  
# mongodb连接uri mongodb://[认证用户名]:[认证密码]@[连接地址]/[额外参数]
uri = 'mongodb://mgouser:mgopass@127.0.0.1:27027/?authSource=admin' 

[mysql]
# mysql连接数据库
database = 'testorm'
# 数据库编码
charset = 'utf8mb4'
# 连接最大生命时间
conn_max_lifetime = '1h'
# mysql连接主机
host = '127.0.0.1'
# mysql连接端口
port = '3306'
# 最大空闲连接数
max_idle_conns = '10'
# 最大连接数
max_open_conns = '100'
# mysql认证用户
user = 'orm'
# mysql认证密码
password = 'password'

[sqlite]
# sqlite数据库文件路径
file_path = './data.db'
# 连接最大生命时间
conn_max_lifetime = '1h'
# 最大空闲连接数
max_idle_conns = '10'
# 最大连接数
max_open_conns = '100'
```

配置文件读取使用了```viper```，所以支持多种配置文件格式，如```json```、```yaml```、```toml```、```ini```等，详情请参考[viper](https://github.com/spf13/viper)文档。

也可以直接传入viper实例进行读取。

# 使用案例
```golang
package main

import (
	"fmt"

	"github.com/lfhy/morm"
)

// 数据库结构体
// 如果数据在mongo需要按mongo-driver进行标注
// 在其他gorm的（mysql，sqlite）按gorm进行标注
type DBSturct struct {
	ID   string `bson:"_id" gorm:"id"`
	Name string `bson:"name" gorm:"name"`
}

// 表名或集合名
func (DBSturct) TableName() string {
	return "dbtable"
}

func main() {
	// 初始化配置文件
	configPath := "/path/to/config.toml"
	err := morm.InitORMConfig(configPath)
	if err != nil {
		fmt.Printf("配置文件加载错误:%v\n", err)
		panic(err)
	}
	// 使用自定义日志:db.SetLogger
	// 数据库初始化
	// 也可以自己处理错误
	// orm, err := morm.InitWithError(configPath)
	// 也可以直接传入配置文件
	// orm := morm.Init(configPath)
	// 如果已经用viper解析了，那也可以使用对应的配置实例 morm.UseViperConfig(viperConfig)
	orm := morm.Init()

	// 创建查询
	var db DBSturct
	db.ID = "123"

	err = orm.Model(&db).Find().One(&db)
	if err != nil {
		fmt.Printf("查询失败:%v\n", err)
		return
	}
	fmt.Printf("查询结果:%+v\n", db)
}

```

# TODO
- 添加测试案例
