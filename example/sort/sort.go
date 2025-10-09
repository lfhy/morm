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
	// 按ID正序
	var idasc, iddesc, createasc, createdesc []*DBSturct
	fmt.Println("查询ID升序数据")
	var sortDB DBSturct
	sortDB.SortDBStruct = SortDBStruct{ID: 1}
	db.M().Where(&DBSturct{IsDelete: 1}).Page(1, 3).Asc(sortDB).All(&idasc)
	for _, data := range idasc {
		fmt.Printf("查询结果:%+v\n", data)
	}
	fmt.Println("查询ID降序数据")
	db.M().Where(&DBSturct{IsDelete: 1}).Page(1, 3).Desc(sortDB).All(&iddesc)
	for _, data := range iddesc {
		fmt.Printf("查询结果:%+v\n", data)
	}
	fmt.Println("查询CreateTime升序数据")
	sortDB.SortDBStruct = SortDBStruct{CreateTime: 1}
	db.M().Where(&DBSturct{IsDelete: 1}).Page(1, 3).Asc(sortDB).All(&createasc)
	for _, data := range createasc {
		fmt.Printf("查询结果:%+v\n", data)
	}
	fmt.Println("查询CreateTime降序数据")
	db.M().Where(&DBSturct{IsDelete: 1}).Page(1, 3).Desc(sortDB).All(&createdesc)
	for _, data := range createdesc {
		fmt.Printf("查询结果:%+v\n", data)
	}
}
