package mongodb

import (
	"context"
	"sync"

	orm "github.com/lfhy/morm/interface"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type CtxKey string

type DBConn struct {
	Database      string //连接的数据库
	NearestClient *mongo.Client
	*mongo.Client
}

var ORMConn *DBConn

type Model struct {
	Tx         DBConn
	Data       any
	OpList     *sync.Map // key:操作模式Mode value:操作值
	WhereList  bson.M
	Ctx        *context.Context //上下文
	Collection string
}

func (m DBConn) Model(data any) orm.ORMModel {
	model := Model{Data: data, Tx: m, WhereList: bson.M{}, OpList: &sync.Map{}}
	model.Collection = model.GetCollection(data)
	return model
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

func (m Model) GetValue(key string) (any, bool) {
	ctx := m.GetContext()
	value := ctx.Value(CtxKey(key))
	return value, value != nil
}

func (m Model) SetValue(key string, value any) orm.ORMModel {
	ctx := m.GetContext()
	ctx = context.WithValue(ctx, CtxKey(key), value)
	m.Ctx = &ctx
	return m
}
