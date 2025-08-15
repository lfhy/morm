package sqlorm

import (
	"context"
	"sync"

	"github.com/lfhy/morm/types"
	"gorm.io/gorm"
)

type Model struct {
	tx     *DBConn
	Data   any
	OpList sync.Map        // key:操作模式Mode value:操作值
	Ctx    context.Context //上下文
}

var ORMConn *DBConn

func (m *DBConn) Model(data any) types.ORMModel {
	return &Model{Data: data, OpList: sync.Map{}, tx: m}
}

func (m *Model) Page(page, limit int) types.ORMModel {
	if page <= 0 {
		page = 1
	}
	return m.Offset((page - 1) * limit).Limit(limit)
}

func (m *Model) Session(transactionFunc func(sessionContext context.Context) error) error {
	return m.tx.Transaction(func(tx *gorm.DB) error {
		return transactionFunc(m.GetContext())
	})
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
