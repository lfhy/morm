package mongodb

import (
	"github.com/lfhy/morm/log"
	"go.mongodb.org/mongo-driver/bson"
)

// Incr 将当前 Where 条件匹配的文档中 field 原子增加 amount。
// 生成 Mongo: {$inc: {field: amount}}
func (m *Model) Incr(column string, amount int64) error {
	m.CheckOID()
	_, err := m.Tx.Client.
		Database(m.Tx.Database).
		Collection(m.GetCollection(m.Data)).
		UpdateMany(m.GetContext(), m.WhereList, bson.M{"$inc": bson.M{column: amount}})
	if err != nil {
		log.Error(err)
	}
	return err
}

// UpdateColumns 用 bson.M / bson.D 原样更新字段。
// data 应为 bson.M 或 bson.D，会作为 $set 传入 UpdateMany。
// 如果 data 本身是带操作符的 bson.M（如 {"$inc": ...}），则原样合并。
func (m *Model) UpdateColumns(data any) error {
	m.CheckOID()
	update := make(bson.M)
	switch v := data.(type) {
	case bson.M:
		// 如果已包含操作符（$ 开头的 key），直接合并
		hasOp := false
		for k := range v {
			if len(k) > 0 && k[0] == '$' {
				hasOp = true
				break
			}
		}
		if hasOp {
			for k, val := range v {
				update[k] = val
			}
		} else {
			update["$set"] = v
		}
	case bson.D:
		update["$set"] = v
	default:
		// 结构体走 ConvertToBSONM
		bsonData, err := ConvertToBSONM(data)
		if err != nil {
			return err
		}
		delete(bsonData, "_id")
		update["$set"] = bsonData
	}
	if len(update) == 0 {
		return nil
	}
	_, err := m.Tx.Client.
		Database(m.Tx.Database).
		Collection(m.GetCollection(m.Data)).
		UpdateMany(m.GetContext(), m.WhereList, update)
	if err != nil {
		log.Error(err)
	}
	return err
}
