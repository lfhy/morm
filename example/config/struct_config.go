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

	// 创建一个数据实例
	dbData := &DBStruct{
		ID:   "123",
		Name: "测试数据",
	}

	// 写入数据
	fmt.Println("正在写入数据...")
	id, err := orm.Model(dbData).Create(&dbData)
	if err != nil {
		fmt.Printf("写入数据失败: %v\n", err)
		return
	}
	fmt.Printf("写入数据成功, ID: %s\n", id)

	// 查询数据
	fmt.Println("正在查询数据...")
	var result DBStruct
	err = orm.Model(&DBStruct{}).Where("id", "123").Find().One(&result)
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}
	fmt.Printf("查询结果: %+v\n", result)

	// 更新数据
	fmt.Println("正在更新数据...")
	dbData.Name = "更新后的数据"
	err = orm.Model(&DBStruct{}).Where("id", "123").Update(dbData)
	if err != nil {
		fmt.Printf("更新失败: %v\n", err)
		return
	}
	fmt.Println("数据更新成功")

	// 再次查询以验证更新
	fmt.Println("正在查询更新后的数据...")
	var updatedResult DBStruct
	err = orm.Model(&DBStruct{}).Where("id", "123").Find().One(&updatedResult)
	if err != nil {
		fmt.Printf("查询更新后的数据失败: %v\n", err)
		return
	}
	fmt.Printf("更新后的查询结果: %+v\n", updatedResult)

	// 删除数据
	fmt.Println("正在删除数据...")
	err = orm.Model(&DBStruct{}).Where("id", "123").Delete()
	if err != nil {
		fmt.Printf("删除失败: %v\n", err)
		return
	}
	fmt.Println("数据删除成功")

	// 尝试查询已删除的数据
	fmt.Println("尝试查询已删除的数据...")
	var deletedResult DBStruct
	err = orm.Model(&DBStruct{}).Where("id", "123").Find().One(&deletedResult)
	if err != nil {
		fmt.Printf("查询已删除数据失败（这是预期的）: %v\n", err)
	} else {
		fmt.Printf("意外: 仍然查询到已删除的数据: %+v\n", deletedResult)
	}
}
