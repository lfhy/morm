package sqlorm

import (
	"context"
	"sync"

	"github.com/lfhy/morm/types"
	"gorm.io/gorm"
)

type Model struct {
	tx           *DBConn
	translatorDB *gorm.DB
	Data         any
	OpList       sync.Map        // key:操作模式Mode value:操作值
	Ctx          context.Context //上下文
	Table        string
}

func (m *Model) getDB() *gorm.DB {
	if m.translatorDB != nil {
		return m.translatorDB
	}
	return m.tx.getDB()
}

var ORMConn *DBConn

func (m *DBConn) Model(data any) types.ORMModel {
	if m.AutoMigrate {
		if m.migrateMap == nil {
			m.migrateMap = make(map[string]bool)
		}
		if _, ok := m.migrateMap[GetTableName(data)]; !ok {
			m.migrateMap[GetTableName(data)] = true
			m.getDB().AutoMigrate(data)
		}
	}
	return &Model{Data: data, OpList: sync.Map{}, tx: m}
}

func (m *Model) Page(page, limit int) types.ORMModel {
	if page <= 0 {
		page = 1
	}
	return m.Offset((page - 1) * limit).Limit(limit)
}

func (m *Model) Session(transactionFunc func(types.Session) error) error {
	return m.tx.Transaction(func(tx *gorm.DB) error {
		if m.translatorDB == nil {
			m.translatorDB = tx
		}
		return transactionFunc(m)
	})
}

func (s *Model) Commit() error {
	return s.getDB().Commit().Error
}

func (s *Model) Rollback() error {
	return s.getDB().Rollback().Error
}

func (m *Model) GetContext() context.Context {
	if m.Ctx != nil {
		return m.Ctx
	} else {
		m.Ctx = context.Background()
	}
	return m.Ctx
}
func (m *Model) SetContext(ctx context.Context) types.ORMModel {
	m.Ctx = ctx
	return m
}
