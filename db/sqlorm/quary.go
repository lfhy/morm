package sqlorm

import (
	"sync"
)

type Quary struct {
	m      Model
	OpList *sync.Map
}

func (q Quary) One(data interface{}) error {
	return q.m.makeQuary().Last(data).Error
}

func (q Quary) All(data interface{}) error {
	return q.m.makeQuary().Find(data).Error
}

func (q Quary) Count() (i int64, err error) {
	return i, q.m.makeQuary().Count(&i).Error
}
