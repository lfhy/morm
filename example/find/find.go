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
	orm := morm.Init()
	if err != nil {
		fmt.Printf("数据库初始化失败:%v\n", err)
		panic(err)
	}

	// 创建查询
	var db DBSturct
	db.ID = "123"

	err = (*orm).Model(&db).Find().One(&db)
	if err != nil {
		fmt.Printf("查询失败:%v\n", err)
		return
	}
	fmt.Printf("查询结果:%+v\n", db)
}
