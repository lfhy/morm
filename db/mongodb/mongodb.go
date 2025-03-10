package mongodb

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"time"

	"github.com/lfhy/morm/conf"
	orm "github.com/lfhy/morm/interface"
	"golang.org/x/net/proxy"

	"github.com/lfhy/morm/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func Init() (orm.ORM, error) {
	ctx := context.Background()
	// 设置连接uri
	opts := options.Client().ApplyURI(conf.ReadConfigToString("mongodb", "uri"))
	proxys := conf.ReadConfigToString("mongodb", "proxy")
	if proxys != "" {
		// socks5://user:pass@host:port
		u, err := url.Parse(proxys)
		if err == nil {
			u.Host = fmt.Sprintf("%s:%s", u.Hostname(), u.Port())
			var auth *proxy.Auth
			if u.User != nil {
				pass, ok := u.User.Password()
				if ok {
					auth = &proxy.Auth{
						User:     u.User.Username(),
						Password: pass,
					}
				}
			}
			dialer, err := proxy.SOCKS5("tcp", u.Host, auth, proxy.Direct)
			if err == nil {
				opts.SetDialer(dialer.(proxy.ContextDialer))
			}
		}
	}
	poolSize := conf.ReadConfigToInt("mongodb", "option_pool_size")
	if poolSize == 0 {
		poolSize = 150
	}

	opts.SetMaxPoolSize(uint64(poolSize))
	opts.SetMinPoolSize(uint64(poolSize / 10))
	// 设置30s 超时
	opts.SetTimeout(30 * time.Second)
	// 只读取主节点
	opts.SetReadPreference(readpref.Primary())
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
	}
	return ""
}

// 启动事务做函数调用
func (m Model) Session(transactionFunc func(SessionContext context.Context) (interface{}, error)) error {
	// 创建会话
	session, err := m.Tx.Client.StartSession()
	if err != nil {
		return log.Error(err)
	}
	defer session.EndSession(context.Background())

	adaptedFunc := func(ctx mongo.SessionContext) (interface{}, error) {
		return transactionFunc(ctx)
	}

	// 使用适配后的函数
	_, err = session.WithTransaction(context.Background(), adaptedFunc)

	if err != nil {
		return log.Error(err)
	}
	return nil
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
			if fieldName == "_id" {
				if _, ok := field.Interface().(primitive.ObjectID); !ok {
					oid2, err := primitive.ObjectIDFromHex(fmt.Sprint(field.Interface()))
					if err != nil {
						bsonData[fieldName] = oid2
						continue
					}
				}
			}
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
	// log.Debugf("data: %+v\n", bsonData)
	result, err := m.Tx.Client.Database(m.Tx.Database).Collection(m.GetCollection(m.Data)).InsertOne(m.GetContext(), bsonData)
	if err != nil {
		log.Error(err)
		return "", err
	}
	id = result.InsertedID.(primitive.ObjectID).Hex()
	setIDField(m.Data, id)
	// log.Debugf("写入后的Date数据: %+v\n", m.Data)
	return id, err
}

// 更新或插入数据
func (m Model) Save(data interface{}) (id string, err error) {
	if data != nil {
		m.Data = data
	}
	bsonData := convertToBSONM(data)

	log.Debugf("MongoDB保存条件: %v\n", bsonData)
	log.Debugf("MongoDB保存Where条件: %v\n", m.WhereList)
	update := bson.M{"$set": bsonData}

	opts := options.Update().SetUpsert(true)
	result, err := m.Tx.Client.Database(m.Tx.Database).Collection(m.GetCollection(m.Data)).UpdateOne(m.GetContext(), m.WhereList, update, opts)
	if err == nil {
		id = fmt.Sprint(result.UpsertedID)
	}
	if id == "" {
		if m.WhereList["_id"] != nil {
			id = fmt.Sprint(m.WhereList["_id"])
		}
	}
	return
}

// 删除
func (m Model) Delete(data interface{}) error {
	if data != nil {
		m.Data = data
	}
	_, err := m.Tx.Client.Database(m.Tx.Database).Collection(m.GetCollection(m.Data)).DeleteMany(m.GetContext(), m.WhereList)
	if err != nil {
		log.Error(err)
	}

	return err
}

// 修改
func (m Model) Update(data interface{}) error {
	if data != nil {
		m.Data = data
	}
	bsonData := convertToBSONM(data)
	log.Debugf("MongoDB更新Update条件: %v\n", bsonData)
	opts := options.Update().SetUpsert(true)
	log.Debugf("MongoDB更新Where条件: %v\n", m.WhereList)
	delete(bsonData, "_id")

	update := bson.M{"$set": bsonData}

	_, err := m.Tx.Client.Database(m.Tx.Database).Collection(m.GetCollection(m.Data)).UpdateMany(m.GetContext(), m.WhereList, update, opts)

	if err != nil {
		log.Error(err)
	}
	return err
}

// 查询数据
func (m Model) Find() orm.ORMQuary {
	return Quary{m: m, Where: m.WhereList}
}
