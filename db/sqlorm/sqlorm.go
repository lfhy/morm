package sqlorm

import (
	"fmt"
	"reflect"
	"sync"

	"gorm.io/gorm"
)

type DBConn struct {
	*gorm.DB
	AutoMigrate bool
	migrateLock sync.RWMutex
	migrateMap  map[string]bool
}

func (m *DBConn) getDB() *gorm.DB {
	return m.DB
}

type Table interface {
	TableName() string
}

// 获取表名
func GetTableName(dest any) string {
	switch v := dest.(type) {
	case Table:
		return v.TableName()
	case *Table:
		return (*v).TableName()
	case string:
		return fmt.Sprint(dest)
	default:
		return reflect.TypeOf(dest).Elem().Name()
	}
}
