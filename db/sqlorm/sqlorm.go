package sqlorm

import (
	"gorm.io/gorm"
)

type DBConn struct {
	*gorm.DB
}

func (m DBConn) getDB() *gorm.DB {
	return m.DB
}
