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

	fmt.Println("查询ID升序数据")
	// 按ID正序
	morm.List(&db,
		&morm.ListOption{
			Page: 1, Limit: 3,
			Sort: []morm.Sort{{
				Key:  &SortDBStruct{ID: 1},
				Mode: morm.Asc,
			},
			}},
		func(m morm.Model) {
			m.Where(&DBSturct{
				IsDelete: 1,
			})
		}, func(data *DBSturct) bool {
			fmt.Printf("查询结果:%+v\n", data)
			return true
		})
	fmt.Println("查询ID降序数据")
	morm.List(&db,
		&morm.ListOption{
			Page: 1, Limit: 3,
			Sort: []morm.Sort{{
				Key:  &SortDBStruct{ID: 1},
				Mode: morm.Desc,
			},
			}},
		func(m morm.Model) {
			m.Where(&DBSturct{
				IsDelete: 1,
			})
		}, func(data *DBSturct) bool {
			fmt.Printf("查询结果:%+v\n", data)
			return true
		})
	fmt.Println("查询CreateTime升序数据")
	morm.List(&db,
		&morm.ListOption{
			Page: 1, Limit: 3,
			Sort: []morm.Sort{{
				Key:  &SortDBStruct{CreateTime: 1},
				Mode: morm.Asc,
			},
			}},
		func(m morm.Model) {
			m.Where(&DBSturct{
				IsDelete: 1,
			})
		}, func(data *DBSturct) bool {
			fmt.Printf("查询结果:%+v\n", data)
			return true
		})
	fmt.Println("查询CreateTime降序数据")
	morm.List(&db,
		&morm.ListOption{
			Page: 1, Limit: 3,
			Sort: []morm.Sort{{
				Key:  &SortDBStruct{CreateTime: 1},
				Mode: morm.Desc,
			},
			}},
		func(m morm.Model) {
			m.Where(&DBSturct{
				IsDelete: 1,
			})
		}, func(data *DBSturct) bool {
			fmt.Printf("查询结果:%+v\n", data)
			return true
		})
	fmt.Println("查询CreateTime升序数据")
	morm.List(&db,
		&morm.ListOption{
			Page: 1, Limit: 3,
			Sort: []morm.Sort{{
				Key:  "create_time",
				Mode: morm.Asc,
			},
			}},
		func(m morm.Model) {
			m.Where(&DBSturct{
				IsDelete: 1,
			})
		}, func(data *DBSturct) bool {
			fmt.Printf("查询结果:%+v\n", data)
			return true
		})
}
