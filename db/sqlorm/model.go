package sqlorm

import (
	"context"
	"database/sql"
	"errors"
	"sync"

	"github.com/lfhy/morm/types"
	"gorm.io/gorm"
)

type Model struct {
	tx                    *DBConn
	translatorDB          *gorm.DB
	userControlTranslator bool
	Data                  any
	OpList                *types.OrderedMap // key:操作模式Mode value:操作值
	upsertOp              sync.Map
	Ctx                   context.Context //上下文
	Table                 string
}

func (m *Model) getDB() *gorm.DB {
	if m.translatorDB != nil {
		return m.translatorDB
	}
	return m.tx.getDB()
}

var ORMConn *DBConn

func (m *DBConn) migrate(data any) {
	if m.migrateMap == nil {
		m.migrateLock.Lock()
		m.migrateMap = make(map[string]bool)
		m.migrateLock.Unlock()
	}
	table := GetTableName(data)
	m.migrateLock.RLock()
	ok := m.migrateMap[table]
	m.migrateLock.RUnlock()
	if !ok {
		m.migrateLock.Lock()
		m.Migrator().AutoMigrate(data)
		m.migrateMap[table] = true
		m.migrateLock.Unlock()
	}
}

func (m *DBConn) Model(data any) types.ORMModel {
	if m.AutoMigrate {
		m.migrate(data)
	}
	return &Model{Data: data, OpList: types.NewOrderedMap(), tx: m, upsertOp: sync.Map{}}
}

func (m *Model) Page(page, limit int) types.ORMModel {
	if page <= 0 {
		page = 1
	}
	return m.Offset((page - 1) * limit).Limit(limit)
}

// SessionModel 是事务期间暴露给用户的 Session 实现。
// 它嵌入 *Model 继承所有 ORMModel 方法，并额外提供 SwitchModel 方法
// 用于在事务内切换到不同的表/结构体，同时共享同一个 gorm 事务。
type sqlSessionModel struct {
	*Model
}

// SwitchModel 返回绑定到当前事务的新 ORMModel，允许跨表操作。
func (s *sqlSessionModel) SwitchModel(data any) types.ORMModel {
	return &Model{
		Data:         data,
		OpList:       types.NewOrderedMap(),
		tx:           s.tx,
		translatorDB: s.translatorDB, // 共享事务 tx
		upsertOp:     sync.Map{},
	}
}

func (m *Model) Session(transactionFunc func(types.Session) error) error {
	err := m.tx.Transaction(func(tx *gorm.DB) error {
		if m.translatorDB == nil {
			m.translatorDB = tx
		}
		return transactionFunc(&sqlSessionModel{Model: m})
	})
	if err != nil && m.userControlTranslator && errors.Is(err, sql.ErrTxDone) {
		return nil
	}
	return err
}

func (s *Model) Commit() error {
	s.userControlTranslator = true
	return s.getDB().Commit().Error
}

func (s *Model) Rollback() error {
	s.userControlTranslator = true
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
