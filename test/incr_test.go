package test

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/lfhy/morm/db/sqlorm"
	"gorm.io/gorm"
)

// incItem 用于测试 Incr / UpdateColumns
type incItem struct {
	ID    int   `gorm:"column:id;primaryKey;autoIncrement"`
	Name  string
	Count int64
}

func (incItem) TableName() string { return "inc_items" }

func newSQLDB(t *testing.T) *sqlorm.DBConn {
	t.Helper()
	// 用文件模式而非 :memory:，确保多连接共享同一数据库（sqlite 内存库每连接独立）
	gdb, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	// 限制单连接，避免 :memory: 的多连接表隔离问题
	if sqlDB, e := gdb.DB(); e == nil {
		sqlDB.SetMaxOpenConns(1)
	}
	if err := gdb.AutoMigrate(&incItem{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return &sqlorm.DBConn{DB: gdb}
}

func TestIncr(t *testing.T) {
	db := newSQLDB(t)

	// 插入初始记录
	item := &incItem{Name: "foo", Count: 10}
	if _, err := db.Model(&incItem{}).Create(item); err != nil {
		t.Fatalf("create: %v", err)
	}

	// 原子自增
	if err := db.Model(&incItem{}).Where("id", item.ID).Incr("count", 5); err != nil {
		t.Fatalf("incr: %v", err)
	}

	// 验证
	var got incItem
	if err := db.Model(&incItem{}).Where("id", item.ID).Find().One(&got); err != nil {
		t.Fatalf("find: %v", err)
	}
	if got.Count != 15 {
		t.Errorf("expected count=15, got %d", got.Count)
	}
}

func TestIncrConcurrent(t *testing.T) {
	db := newSQLDB(t)
	item := &incItem{Name: "bar", Count: 0}
	if _, err := db.Model(&incItem{}).Create(item); err != nil {
		t.Fatalf("create: %v", err)
	}

	// 并发自增 100 次，每次 +1
	done := make(chan struct{}, 100)
	for i := 0; i < 100; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			db.Model(&incItem{}).Where("id", item.ID).Incr("count", 1)
		}()
	}
	for i := 0; i < 100; i++ {
		<-done
	}

	var got incItem
	if err := db.Model(&incItem{}).Where("id", item.ID).Find().One(&got); err != nil {
		t.Fatalf("find: %v", err)
	}
	if got.Count != 100 {
		t.Errorf("expected count=100 after 100 concurrent increments, got %d", got.Count)
	}
}

func TestUpdateColumns(t *testing.T) {
	db := newSQLDB(t)
	item := &incItem{Name: "baz", Count: 1}
	if _, err := db.Model(&incItem{}).Create(item); err != nil {
		t.Fatalf("create: %v", err)
	}

	// UpdateColumns 用 map 原样更新（不跳过零值）
	if err := db.Model(&incItem{}).Where("id", item.ID).UpdateColumns(map[string]any{"count": 0, "name": "updated"}); err != nil {
		t.Fatalf("updateColumns: %v", err)
	}

	var got incItem
	if err := db.Model(&incItem{}).Where("id", item.ID).Find().One(&got); err != nil {
		t.Fatalf("find: %v", err)
	}
	if got.Count != 0 || got.Name != "updated" {
		t.Errorf("expected count=0 name=updated, got count=%d name=%s", got.Count, got.Name)
	}
}
