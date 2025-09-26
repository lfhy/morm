package main

import (
	"fmt"

	"github.com/lfhy/morm"
	"github.com/lfhy/morm/conf"
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

func (DBSturct) M() morm.ORMModel {
	return db.Model(&DBSturct{})
}

var db morm.ORM

func init() {
	// // 初始化配置文件
	// configPath := "/path/to/config.toml"
	// err := morm.InitORMConfig(configPath)
	// if err != nil {
	// 	fmt.Printf("配置文件加载错误:%v\n", err)
	// 	panic(err)
	// }
	// // 使用自定义日志:db.SetLogger
	// // 数据库初始化
	// db = morm.Init()

	// 通过结构体配置初始化
	dbConfig := &conf.DBConfig{
		Type: "sqlite",
		SQLiteConfig: &conf.SQLiteConfig{
			AutoCreateTable: true,
			FilePath:        "file:testdatabase?mode=memory&cache=shared",
			ConnMaxLifetime: "1h",
			MaxIdleConns:    "10",
			MaxOpenConns:    "100",
		},
	}

	// 使用配置结构体初始化ORM
	db = morm.InitWithDBConfig(dbConfig)
}

func main() {
	var db DBSturct

	// 写入数据
	db.Name = "test"
	id, err := db.M().Create(&db)
	if err != nil {
		fmt.Printf("写入数据失败:%v\n", err)
		return
	}
	fmt.Printf("写入数据成功,id:%s\n", id)

	// 创建查询
	var find DBSturct
	err = db.M().Where("id", db.ID).Find().One(&find)
	if err != nil {
		fmt.Printf("查询失败:%v\n", err)
		return
	}
	fmt.Printf("查询结果:%+v\n", find)

	// 更新数据
	err = db.M().Where(&DBSturct{ID: db.ID}).Update(&DBSturct{Name: "test1"})
	if err != nil {
		fmt.Printf("更新失败:%v\n", err)
		return
	}
	err = db.M().Where("id", db.ID).Find().One(&find)
	if err != nil {
		fmt.Printf("查询失败:%v\n", err)
		return
	}
	fmt.Printf("更新结果:%+v\n", find)

	// 删除数据
	err = db.M().Where("id", db.ID).Delete()
	if err != nil {
		fmt.Printf("删除失败:%v\n", err)
		return
	}
	fmt.Println("删除成功")
}
