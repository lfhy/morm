package main

import (
	"fmt"

	"github.com/lfhy/morm"
	"github.com/lfhy/morm/conf"
)

type User struct {
	ID        string `bson:"_id" gorm:"column:id"`
	Name      string `bson:"name" gorm:"column:name"`
	Age       int    `bson:"age" gorm:"column:age"`
	CreatedAt int64  `bson:"created_at" gorm:"column:created_at"`
}

func (User) TableName() string {
	return "user"
}

func main() {
	// 通过结构体配置初始化
	dbConfig := &conf.DBConfig{
		Type: "sqlite",
		LogConfig: &conf.LogConfig{
			Log:      "./db.log",
			LogLevel: morm.LogLevelInfo,
		},
		SQLiteConfig: &conf.SQLiteConfig{
			AutoCreateTable: true,
			FilePath:        "./test.db",
			ConnMaxLifetime: "1h",
			MaxIdleConns:    "10",
			MaxOpenConns:    "100",
		},
	}
	// 使用配置结构体初始化ORM
	orm := morm.InitWithDBConfig(dbConfig)

	// 插入一条数据
	orm.Model(&User{}).Create(&User{Name: "test", Age: 18})

	// 查询一条数据
	var userold User
	orm.Model(&User{}).Where(&User{Name: "test"}).Find().One(&userold)

	fmt.Printf("原始数据:%v\n", userold)

	// 测试事务回滚
	err := orm.Model(&User{}).Session(func(sessionModel morm.Session) error {
		sessionModel.Where(&User{Name: "test"}).Update(&User{Age: 19})
		var user User
		sessionModel.Where(&User{Name: "test"}).Find().One(&user)
		fmt.Printf("事务中的数据: %v\n", user)
		return sessionModel.Rollback()
	})
	fmt.Printf("事务处理错误: %v\n", err)
	// 回滚后查询数据
	var user User
	orm.Model(&User{}).Where(&User{Name: "test"}).Find().One(&user)
	fmt.Printf("事务回滚后的数据: %v\n", user)

	// 测试事务提交
	err = orm.Model(&User{}).Session(func(sessionModel morm.Session) error {
		sessionModel.Where(&User{Name: "test"}).Update(&User{Age: 20})
		var user User
		sessionModel.Where(&User{Name: "test"}).Find().One(&user)
		fmt.Printf("事务中的数据: %v\n", user)
		return sessionModel.Commit()
	})

	fmt.Printf("事务处理错误: %v\n", err)

	// 提交后的数据
	var user2 User
	orm.Model(&User{}).Where(&User{Name: "test"}).Find().One(&user2)
	fmt.Printf("事务提交后的数据: %v\n", user2)
}
