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
	ID   string `bson:"_id" gorm:"id"`
	Name string `bson:"name" gorm:"name"`
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

	// 写入数据
	fmt.Println("开始写入数据...")
	for i := 0; i < 100; i++ {
		dbData := &DBStruct{
			ID:   fmt.Sprint(i),
			Name: "测试数据" + fmt.Sprint(i),
		}
		_, err := orm.Model(dbData).Create(dbData)
		if err != nil {
			fmt.Printf("写入数据失败: %v\n", err)
			return
		}
	}
	fmt.Println("数据写入成功")
	// 使用游标查询
	cursor, err := orm.Model(&DBStruct{}).Find().Cursor()
	if err != nil {
		fmt.Printf("使用游标查询失败: %v\n", err)
		return
	}
	defer cursor.Close()
	var index int
	for cursor.Next() {
		var dbData DBStruct
		err := cursor.Decode(&dbData)
		if err != nil {
			fmt.Printf("解码数据失败: %v\n", err)
			return
		}
		fmt.Printf("查询结果: %+v\n", dbData)
		index++
	}
	fmt.Printf("查询到%d条数据\n", index)

}
