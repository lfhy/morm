# Morm

Mixed ORM that allows easy access to different types of databases by configuring data sources. Currently supports ```mysql```, ```mongodb``` and ```sqlite```, implemented using ```gorm``` and ```mongo-driver```.

Configuration files are parsed using ```viper```.

[![zread](https://img.shields.io/badge/Ask_Zread-_.svg?style=flat&color=00b0aa&labelColor=000000&logo=data%3Aimage%2Fsvg%2Bxml%3Bbase64%2CPHN2ZyB3aWR0aD0iMTYiIGhlaWdodD0iMTYiIHZpZXdCb3g9IjAgMCAxNiAxNiIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHBhdGggZD0iTTQuOTYxNTYgMS42MDAxSDIuMjQxNTZDMS44ODgxIDEuNjAwMSAxLjYwMTU2IDEuODg2NjQgMS42MDE1NiAyLjI0MDFWNC45NjAxQzEuNjAxNTYgNS4zMTM1NiAxLjg4ODEgNS42MDAxIDIuMjQxNTYgNS42MDAxSDQuOTYxNTZDNS4zMTUwMiA1LjYwMDEgNS42MDE1NiA1LjMxMzU2IDUuNjAxNTYgNC45NjAxVjIuMjQwMUM1LjYwMTU2IDEuODg2NjQgNS4zMTUwMiAxLjYwMDEgNC45NjE1NiAxLjYwMDFaIiBmaWxsPSIjZmZmIi8%2BCjxwYXRoIGQ9Ik00Ljk2MTU2IDEwLjM5OTlIMi4yNDE1NkMxLjg4ODEgMTAuMzk5OSAxLjYwMTU2IDEwLjY4NjQgMS42MDE1NiAxMS4wMzk5VjEzLjc1OTlDMS42MDE1NiAxNC4xMTM0IDEuODg4MSAxNC4zOTk5IDIuMjQxNTYgMTQuMzk5OUg0Ljk2MTU2QzUuMzE1MDIgMTQuMzk5OSA1LjYwMTU2IDE0LjExMzQgNS42MDE1NiAxMy43NTk5VjExLjAzOTlDNS42MDE1NiAxMC42ODY0IDUuMzE1MDIgMTAuMzk5OSA0Ljk2MTU2IDEwLjM5OTlaIiBmaWxsPSIjZmZmIi8%2BCjxwYXRoIGQ9Ik0xMy43NTg0IDEuNjAwMUgxMS4wMzg0QzEwLjY4NSAxLjYwMDEgMTAuMzk4NCAxLjg4NjY0IDEwLjM5ODQgMi4yNDAxVjQuOTYwMUMxMC4zOTg0IDUuMzEzNTYgMTAuNjg1IDUuNjAwMSAxMS4wMzg0IDUuNjAwMUgxMy43NTg0QzE0LjExMTkgNS42MDAxIDE0LjM5ODQgNS4zMTM1NiAxNC4zOTg0IDQuOTYwMVYyLjI0MDFDMTQuMzk4NCAxLjg4NjY0IDE0LjExMTkgMS42MDAxIDEzLjc1ODQgMS42MDAxWiIgZmlsbD0iI2ZmZiIvPgo8cGF0aCBkPSJNNCAxMkwxMiA0TDQgMTJaIiBmaWxsPSIjZmZmIi8%2BCjxwYXRoIGQ9Ik00IDEyTDEyIDQiIHN0cm9rZT0iI2ZmZiIgc3Ryb2tlLXdpZHRoPSIxLjUiIHN0cm9rZS1saW5lY2FwPSJyb3VuZCIvPgo8L3N2Zz4K&logoColor=ffffff)](https://zread.ai/lfhy/morm)

# Initialization via Configuration Struct

In addition to initializing the ORM with a configuration file, you can also initialize the ORM directly through the DBConfig struct:

```golang
package main

import (
	"github.com/lfhy/morm"
	"github.com/lfhy/morm/conf"
)

func main() {
	// Initialize via struct configuration
	dbConfig := &conf.DBConfig{
		Type: "sqlite",
		LogConfig: &conf.LogConfig{
			Log:      "./db.log",
			LogLevel: "4",
		},
		SQLiteConfig: &conf.SQLiteConfig{
			AutoCreateTable: true,
			FilePath:        "./test.db",
			ConnMaxLifetime: "1h",
			MaxIdleConns:    "10",
			MaxOpenConns:    "100",
		},
	}

	// Initialize ORM with configuration struct
	orm := morm.InitWithDBConfig(dbConfig)
	
	// Use ORM for database operations...
}
```

# Configuration File Reference

```toml
[db]
log = './db.log'    # Log file path
loglevel = '4'  # Log level 
type = 'mysql' # Default ORM type

[mongodb]
# Database to connect to for mongodb
database = 'testorm'    
# Connection pool size
option_pool_size = '200'   
# Mongodb proxy connection method
proxy = 'socks5://127.0.0.1:7890'  
# Mongodb connection URI mongodb://[username]:[password]@[address]/[parameters]
uri = 'mongodb://mgouser:mgopass@127.0.0.1:27027/?authSource=admin' 

[mysql]
# MySQL database to connect to
database = 'testorm'
# Database encoding
charset = 'utf8mb4'
# Connection maximum lifetime
conn_max_lifetime = '1h'
# MySQL connection host
host = '127.0.0.1'
# MySQL connection port
port = '3306'
# Maximum idle connections
max_idle_conns = '10'
# Maximum open connections
max_open_conns = '100'
# MySQL authentication user
user = 'orm'
# MySQL authentication password
password = 'password'

[sqlite]
# SQLite database file path
file_path = './data.db'
# Connection maximum lifetime
conn_max_lifetime = '1h'
# Maximum idle connections
max_idle_conns = '10'
# Maximum open connections
max_open_conns = '100'
```

Configuration file reading uses ```viper```, so multiple configuration file formats are supported, such as ```json```, ```yaml```, ```toml```, ```ini```, etc. For details, please refer to [viper](https://github.com/spf13/viper) documentation.

You can also pass in a viper instance for reading.

# Usage Examples

```golang
package main

import (
	"fmt"

	"github.com/lfhy/morm"
)

// Database struct
// If data is in mongo, annotate according to mongo-driver
// In other gorm databases (mysql, sqlite), annotate according to gorm
type DBSturct struct {
	ID   string `bson:"_id" gorm:"id"`
	Name string `bson:"name" gorm:"name"`
}

// Table name or collection name
func (DBSturct) TableName() string {
	return "dbtable"
}

func main() {
	// Initialize configuration file
	configPath := "/path/to/config.toml"
	err := morm.InitORMConfig(configPath)
	if err != nil {
		fmt.Printf("Configuration file loading error:%v\n", err)
		panic(err)
	}
	// Use custom logger: db.SetLogger
	// Database initialization
	// You can also handle errors yourself
	// orm, err := morm.InitWithError(configPath)
	// You can also pass the configuration file directly
	// orm := morm.Init(configPath)
	// If you've already parsed with viper, you can also use the corresponding configuration instance morm.UseViperConfig(viperConfig)
	orm := morm.Init()

	// Create query
	var db DBSturct
	db.ID = "123"

	err = orm.Model(&db).Find().One(&db)
	if err != nil {
		fmt.Printf("Query failed:%v\n", err)
		return
	}
	fmt.Printf("Query result:%+v\n", db)
}

```

# TODO
- Add test cases