package main

import (
	"fmt"
	"time"

	"github.com/lfhy/morm"
	"github.com/lfhy/morm/conf"
)

var (
	StatusWait  = 0
	StatusDoing = 1
	StatusDone  = 2
)

var (
	PriorityLow  = 1
	PriorityMid  = 2
	PriorityHigh = 3
)

var (
	IsNotDeleted = 0
	IsDeleted    = 1
)

// Task 表结构
type Task struct {
	ID         int    `bson:"_id" gorm:"id"`
	Title      string `bson:"title" gorm:"title"`
	Priority   int    `bson:"priority" gorm:"priority"`       // 优先级：1-低，2-中，3-高
	Status     *int   `bson:"status" gorm:"status"`           // 状态：0-未开始，1-进行中，2-已完成
	Deadline   int64  `bson:"deadline" gorm:"deadline"`       // 时间戳
	CreateTime int64  `bson:"create_time" gorm:"create_time"` // 时间戳
	IsDeleted  *int   `bson:"is_deleted" gorm:"is_deleted"`   // 是否删除
}

func (s Task) String() string {
	return fmt.Sprintf("任务ID: %d, 标题: %s, 优先级: %d, 状态: %d, 截止时间: %s, 创建时间: %s, 是否删除: %d",
		s.ID, s.Title, s.Priority, *s.Status, time.Unix(s.Deadline, 0).Format("2006-01-02 15:04:05"), time.Unix(s.CreateTime, 0).Format("2006-01-02 15:04:05"), *s.IsDeleted,
	)
}

func (Task) TableName() string {
	return "tasks"
}

func (Task) M() morm.ORMModel {
	return db.Model(&Task{})
}

var db morm.ORM

func init() {
	dbConfig := &conf.DBConfig{
		Type: "sqlite",
		SQLiteConfig: &conf.SQLiteConfig{
			AutoCreateTable: true,
			FilePath:        "file:taskdb?mode=memory&cache=shared",
			ConnMaxLifetime: "1h",
			MaxIdleConns:    "10",
			MaxOpenConns:    "100",
		},
	}
	db = morm.InitWithDBConfig(dbConfig)
}

// 写入测试数据
func insertTestData() {
	tasks := []Task{
		{Title: "写报告", Priority: PriorityHigh, Status: &StatusDoing, Deadline: time.Now().Add(24 * time.Hour).Unix(), CreateTime: time.Now().Unix(), IsDeleted: &IsNotDeleted},
		{Title: "开会准备", Priority: PriorityMid, Status: &StatusWait, Deadline: time.Now().Add(6 * time.Hour).Unix(), CreateTime: time.Now().Add(-3 * time.Hour).Unix(), IsDeleted: &IsNotDeleted},
		{Title: "修复bug", Priority: PriorityHigh, Status: &StatusDoing, Deadline: time.Now().Add(6 * time.Hour).Unix(), CreateTime: time.Now().Add(-2 * time.Hour).Unix(), IsDeleted: &IsNotDeleted},
		{Title: "完成项目", Priority: PriorityLow, Status: &StatusDone, Deadline: time.Now().Add(-24 * time.Hour).Unix(), CreateTime: time.Now().Add(-3 * time.Hour).Unix(), IsDeleted: &IsNotDeleted},
		{Title: "准备材料", Priority: PriorityMid, Status: &StatusWait, Deadline: time.Now().Add(6 * time.Hour).Unix(), CreateTime: time.Now().Add(-1 * time.Hour).Unix(), IsDeleted: &IsNotDeleted},
		{Title: "更新文档", Priority: PriorityMid, Status: &StatusWait, Deadline: time.Now().Add(6 * time.Hour).Unix(), CreateTime: time.Now().Add(-5 * time.Hour).Unix(), IsDeleted: &IsDeleted}, // 已删除
	}

	for _, task := range tasks {
		id, err := db.Model(&Task{}).Create(&task)
		if err != nil {
			panic(err)
		}
		fmt.Println("插入任务成功, ID:", id)
	}
}

func main() {
	insertTestData()

	fmt.Println("\n=== 复杂排序查询 ===")
	// 复杂排序：
	// 1. 只显示未删除的任务
	// 2. 先显示 status < 2 的任务
	// 3. 再按 priority DESC, deadline ASC, create_time ASC 排序
	// 预期结果：
	// 开会准备（priority=2, deadline 最早）
	// 准备材料（priority=2, deadline 同上，但 create_time 更早）
	// 修复bug（priority=3, deadline 更近）
	// 写报告（priority=3, deadline 较远）
	morm.List(&Task{},
		&morm.ListOption{
			Page: 1, Limit: 10,
			Sorts: []*morm.Sort{
				{Key: "status", Mode: morm.Asc},      // 优先处理 status < 2 的任务
				{Key: "priority", Mode: morm.Desc},   // 优先级高的排前面
				{Key: "deadline", Mode: morm.Asc},    // 截止时间早的排前面
				{Key: "create_time", Mode: morm.Asc}, // 创建时间早的排前面
			},
		},
		func(m morm.Model) {
			m.Where(&Task{IsDeleted: &IsNotDeleted}).WhereLt(&Task{Status: &StatusDone})
		},
		func(data *Task) bool {
			fmt.Printf("查询结果: %s\n", data)
			return true
		})
}
