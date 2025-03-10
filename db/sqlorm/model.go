package sqlorm

import (
	"context"
	"errors"
	"sync"

	orm "github.com/lfhy/morm/interface"
)

type Model struct {
	tx     DBConn
	Data   interface{}
	OpList *sync.Map        // key:操作模式Mode value:操作值
	Ctx    *context.Context //上下文
}

var ORMConn *DBConn

func (m DBConn) Model(data interface{}) orm.ORMModel {
	return Model{Data: data, OpList: &sync.Map{}, tx: m}
}

func (m Model) Page(page, limit int) orm.ORMModel {
	if page <= 0 {
		page = 1
	}
	return m.Offset((page - 1) * limit).Limit(limit)
}
func (m Model) Session(transactionFunc func(sessionContext context.Context) (interface{}, error)) error {
	return errors.New("方法未实现")
}
func (m Model) GetContext() context.Context {
	if m.Ctx != nil {
		return *m.Ctx
	} else {
		*m.Ctx = context.Background()
	}
	return *m.Ctx
}
func (m Model) SetContext(ctx context.Context) orm.ORMModel {
	m.Ctx = &ctx
	return m
}
