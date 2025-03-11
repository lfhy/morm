package mongodb

import (
	"context"

	"github.com/lfhy/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Quary struct {
	m     Model
	Where bson.M
}

func (q Quary) One(data any) error {
	log.Debugf("查询集合 %v ,Mongo查询条件: %+v \n", q.m.GetCollection(q.m.Data), q.m.WhereList)
	result := q.m.Tx.Client.Database(q.m.Tx.Database).Collection(q.m.GetCollection(q.m.Data)).FindOne(q.m.GetContext(), q.m.WhereList)
	err := result.Decode(data)
	if err != nil {
		log.Warnf("查询集合 %v ,Mongo查询条件: %+v 错误: %v\n", q.m.GetCollection(q.m.Data), q.m.WhereList, err)
	}
	log.Debugf("Mongo查询结果: %+v\n", data)

	return err
}

// 查询全部
func (q Quary) All(data any) error {
	opts := q.m.makeQuary()
	log.Debugf("查询集合 %v Mongo查询条件: %v\n", q.m.GetCollection(q.m.Data), q.m.WhereList)
	log.Debugf("Mongo查询限制: %+v\n", opts)
	result, err := q.m.Tx.Client.Database(q.m.Tx.Database).Collection(q.m.GetCollection(q.m.Data)).Find(q.m.GetContext(), q.m.WhereList, &opts)

	// log.Debugf("Mongo查询结果: %+v\n", result)
	if err != nil {
		log.Warnf("Mongo查出错: %v\n", err)
		return err
	}
	err = result.All(context.Background(), data)
	if err != nil {
		log.Warnf("mongdob查询数据ALL Decode失败: %v\n", err)
	}
	return err
}

func (q Quary) Count() (int64, error) {
	return q.m.Tx.Client.Database(q.m.Tx.Database).Collection(q.m.GetCollection(q.m.Data)).CountDocuments(q.m.GetContext(), q.m.WhereList)
}

type IDModel struct {
	ID primitive.ObjectID `bson:"_id"`
}

// 删除查询结果
func (q Quary) Delete() error {
	var deleteIDs []*IDModel
	err := q.All(&deleteIDs)
	if err != nil {
		return err
	}
	// 批量删除
	var models []mongo.WriteModel
	for _, id := range deleteIDs {
		models = append(models, mongo.NewDeleteOneModel().SetFilter(id))
	}
	// 执行批量写入操作
	bulkWriteOpts := options.BulkWrite().SetOrdered(false) // 设置为无序以提高性能
	_, err = q.m.Tx.Client.Database(q.m.Tx.Database).Collection(q.m.GetCollection(q.m.Data)).BulkWrite(q.m.GetContext(), models, bulkWriteOpts)
	return err
}
