package mongodb

import (
	"context"
	"fmt"
	"reflect"

	orm "github.com/lfhy/morm"

	"github.com/lfhy/morm/conf"
	"github.com/lfhy/morm/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Init() (orm.ORM, error) {
	ctx := context.Background()
	// 设置连接uri
	opts := options.Client().ApplyURI(conf.ReadConfigToString("mongodb", "uri"))
	// 连接mongodb
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}
	conn := DBConn{Database: conf.ReadConfigToString("mongodb", "database"), Client: client}
	ORMConn = &conn
	return ORMConn, nil
}

// 获取集合
func (m Model) GetCollection(dest interface{}) string {
	switch v := dest.(type) {
	case Table:
		return v.TableName()
	case *Table:
		return (*v).TableName()
	default:
		return fmt.Sprint(v)
	}
}

type Table interface {
	TableName() string
}

// 转换data TO bsonM
func convertToBSONM(data interface{}) bson.M {
	bsonData := bson.M{}
	val := reflect.ValueOf(data)

	// Check if the value is a pointer, and if so, get the underlying element
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldName := typ.Field(i).Tag.Get("bson")
		// Skip fields with zero values
		if !fieldIsZero(field) {
			bsonData[fieldName] = field.Interface()
		}
	}

	return bsonData
}

func fieldIsZero(field reflect.Value) bool {
	zeroValue := reflect.Zero(field.Type())
	return reflect.DeepEqual(field.Interface(), zeroValue.Interface())
}

func (m Model) Create(data interface{}) (id string, err error) {
	if data != nil {
		m.Data = data
	}
	bsonData := convertToBSONM(m.Data)

	log.Debugf("MongoCreate", "data: %+v", bsonData)
	result, err := m.Tx.Client.Database(m.Tx.Database).Collection(m.GetCollection(m.Data)).InsertOne(context.Background(), bsonData)
	if err != nil {
		return "", log.Error(err)
	}
	id = result.InsertedID.(primitive.ObjectID).Hex()
	if err == nil {
		setIDField(m.Data, id)
	}
	log.Debugf("MongoCreate", "写入后的Date数据: %+v", m.Data)
	return id, err
}

// 更新或插入数据
func (m Model) Save(data interface{}) (id string, err error) {
	if data != nil {
		m.Data = data
	}
	opts := options.Update().SetUpsert(true)
	result, err := m.Tx.Client.Database(m.Tx.Database).Collection(m.GetCollection(m.Data)).UpdateOne(context.Background(), m.WhereList, data, opts)
	if err == nil {
		id = fmt.Sprint(result.UpsertedID)
	}
	if id == "" {
		id = fmt.Sprint(result.UpsertedID)
	}
	return
}

// 删除
func (m Model) Delete(data interface{}) error {
	if data != nil {
		m.Data = data
	}
	log.Debugf("MongoDelete", "m.WhereList: %v", m.WhereList)
	result, err := m.Tx.Client.Database(m.Tx.Database).Collection(m.GetCollection(m.Data)).DeleteMany(context.Background(), m.WhereList)
	if err != nil {
		return log.Error(err)
	}
	log.Debugf("MongoDelete", "result: %v", result)

	return nil
}

// 修改
func (m Model) Update(data interface{}) error {
	if data != nil {
		m.Data = data
	}
	bsonData := convertToBSONM(data)
	log.Debugf("MongoUpdate", "MongoDB更新Update条件: %v", bsonData)
	opts := options.Update().SetUpsert(false)
	log.Debugf("MongoUpdate", "MongoDB更新Where条件: %v", m.WhereList)
	update := bson.M{"$set": bsonData}

	_, err := m.Tx.Client.Database(m.Tx.Database).Collection(m.GetCollection(m.Data)).UpdateMany(context.Background(), m.WhereList, update, opts)

	if err != nil {
		log.Error(err)
	}
	return err
}

// 查询数据
func (m Model) Find() orm.ORMQuary {
	return Quary{m: m, Where: m.WhereList}
}
