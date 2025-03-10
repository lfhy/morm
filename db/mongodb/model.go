package mongodb

import (
	"context"
	"sync"

	orm "github.com/lfhy/morm/interface"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type DBConn struct {
	Database string //连接的数据库

	*mongo.Client
}

var ORMConn *DBConn

type Model struct {
	Tx        DBConn
	Data      interface{}
	OpList    *sync.Map // key:操作模式Mode value:操作值
	WhereList bson.M
	Ctx       *context.Context //上下文

}

func (m DBConn) Model(data interface{}) orm.ORMModel {
	return Model{Data: data, Tx: m, WhereList: bson.M{}, OpList: &sync.Map{}}
}

func (m Model) Page(page, limit int) orm.ORMModel {
	if page <= 0 {
		page = 1
	}
	return m.Offset((page - 1) * limit).Limit(limit)
}
func (m Model) GetContext() context.Context {
	if m.Ctx != nil {
		return *m.Ctx
	} else {
		ctx := context.Background()
		m.Ctx = &ctx
	}
	return *m.Ctx
}
func (m Model) SetContext(ctx context.Context) orm.ORMModel {
	m.Ctx = &ctx
	return m
}
