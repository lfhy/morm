package mongodb

import (
	"context"
	"fmt"

	"github.com/lfhy/morm/log"
	"go.mongodb.org/mongo-driver/bson"
)

type Quary struct {
	m     Model
	Where bson.M
}

func (q Quary) One(data interface{}) error {
	log.Debugf("MongoFindOne", "查询集合 %v ,Mongo查询条件: %+v", q.m.GetCollection(q.m.Data), q.m.WhereList)
	result := q.m.Tx.Client.Database(q.m.Tx.Database).Collection(q.m.GetCollection(q.m.Data)).FindOne(context.Background(), q.m.WhereList)
	fmt.Printf("result.Err(): %v\n", result.Err())
	err := result.Decode(data)
	if err != nil {
		return log.Error(err)
	}
	log.Debugf("Mongo查询结果: %+v\n", data)

	return err
}
func (q Quary) All(data interface{}) error {
	opts := q.m.makeQuary()
	log.Debugf("Mongo查询条件: %+v\n", q.m.WhereList)
	log.Debugf("Mongo查询限制: %+v\n", opts)
	result, err := q.m.Tx.Client.Database(q.m.Tx.Database).Collection(q.m.GetCollection(q.m.Data)).Find(context.Background(), q.m.WhereList, &opts)
	log.Debugf("Mongo查询结果: %+v\n", result)
	if err != nil {
		return log.Errorf("%v", err)
	}
	err = result.All(context.Background(), data)
	if err != nil {
		return log.Errorf("%v", err)
	}
	return err
}

func (q Quary) Count() (i int64, err error) {
	log.Debugf("Mongo查询条件: %v\n", q.m.WhereList)
	count, err := q.m.Tx.Client.Database(q.m.Tx.Database).Collection(q.m.GetCollection(q.m.Data)).CountDocuments(context.Background(), q.m.WhereList)
	return count, err
}
