package main

import (
	"fmt"

	"github.com/lfhy/morm"
	"github.com/lfhy/morm/conf"
)

// 数据库结构体
// 如果数据在mongo需要按mongo-driver进行标注
// 在其他gorm的（mysql，sqlite）按gorm进行标注
type DBStruct struct {
	ID    string `bson:"_id" gorm:"id"`
	Name  string `bson:"name" gorm:"name"`
	Value int    `bson:"value" gorm:"value"`
}

// 表名或集合名
func (DBStruct) TableName() string {
	return "dbtable"
}

func main() {
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
	orm := morm.InitWithDBConfig(dbConfig)

	fmt.Println("第一次写入数据")
	err := orm.Model(&DBStruct{}).Where(&DBStruct{Name: "test"}).Save(&DBStruct{
		Value: 1,
	})
	if err != nil {
		fmt.Println("写入数据失败:", err)
		return
	}
	// 查询写入的数据
	var dbData DBStruct
	err = orm.Model(&DBStruct{}).Where(&DBStruct{Name: "test"}).One(&dbData)
	if err != nil {
		fmt.Println("查询数据失败:", err)
		return
	}
	fmt.Println("查询结果:", dbData)
}
