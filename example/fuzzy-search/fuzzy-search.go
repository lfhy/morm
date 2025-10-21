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
	SortDBStruct
	Name     string `bson:"name" gorm:"name"`
	IsDelete int    `bson:"is_delete" gorm:"is_delete"`
}

type SortDBStruct struct {
	ID         int `bson:"_id" gorm:"id"`
	CreateTime int `bson:"create_time" gorm:"create_time"`
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

func isDelete(index int) int {
	if index%2 == 0 {
		return 1
	}
	return 2
}

func main() {
	// 写入多个数据
	var db DBSturct
	for i := 0; i < 10; i++ {
		id, err := db.M().Create(&DBSturct{Name: "test" + fmt.Sprint(i), SortDBStruct: SortDBStruct{CreateTime: 10 - i}, IsDelete: isDelete(i)})
		if err != nil {
			panic(err)
		}
		println("写入数据成功,id:", id)
	}
	// 模糊查询
	var dbs []*DBSturct
	db.M().WhereLike(&DBSturct{Name: "est1"}).Find().All(&dbs)
	for _, data := range dbs {
		fmt.Printf("data: %+v\n", *data)
	}
}
