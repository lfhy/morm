package sqlorm

import (
	"sync"

	orm "github.com/lfhy/morm/interface"
)

type Model struct {
	tx     DBConn
	Data   interface{}
	OpList *sync.Map // key:操作模式Mode value:操作值
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
