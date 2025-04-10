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
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
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
	readMode := conf.ReadConfigToString("mongodb", "readmode")
	W := conf.ReadConfigToInt("mongodb", "w")
	if W == 0 {
		//只写到主节点
		W = 1
	}
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
	// // 写确认
	opts.SetWriteConcern(&writeconcern.WriteConcern{
		W:        W,
		WTimeout: 1 * time.Second,
	})
	// 连接mongodb
	client, err := mongo.Connect(ctx, opts)

	if err != nil {
		return nil, err
	}
	conn := DBConn{Database: conf.ReadConfigToString("mongodb", "database"), Client: client, NearestClient: client}
	ORMConn = &conn
	if readMode == "master" {
		return ORMConn, nil
	}

	// 使用就近读取
	opts.SetReadPreference(readpref.Nearest(
		readpref.WithHedgeEnabled(true),
		readpref.WithMaxStaleness(5*time.Minute),
	))
	// 连接mongodb
	nearestClient, err := mongo.Connect(ctx, opts)
	if err != nil {
		client.Disconnect(ctx)
		return nil, err
	}
	conn.NearestClient = nearestClient

	return ORMConn, nil
}

// 获取集合
func (m Model) GetCollection(dest any) string {
	if m.Collection != "" {
		return m.Collection
	}
	switch v := dest.(type) {
	case Table:
		m.Collection = v.TableName()
	case *Table:
		m.Collection = (*v).TableName()
	case string:
		m.Collection = fmt.Sprint(dest)
	default:
		m.Collection = reflect.TypeOf(dest).Elem().Name()
	}
	return m.Collection
}

// 启动事务做函数调用
func (m Model) Session(transactionFunc func(SessionContext context.Context) error) error {
	// 创建会话
	session, err := m.Tx.Client.StartSession()
	if err != nil {
		return log.Error(err)
	}
	// 使用适配后的函数
	sctx := m.GetContext()
	defer session.EndSession(sctx)

	adaptedFunc := func(ctx mongo.SessionContext) (any, error) {
		return nil, transactionFunc(ctx)
	}

	_, err = session.WithTransaction(sctx, adaptedFunc)
	if err != nil {
		session.AbortTransaction(sctx)
		return log.Error(err)
	}
	session.CommitTransaction(sctx)
	return nil
}

type Table interface {
	TableName() string
}

// 转换data TO bsonM
func convertToBSONM(data any) bson.M {
	bsonData := bson.M{}
	val := reflect.ValueOf(data)
	// Check if the value is a pointer, and if so, get the underlying element
	for val.Kind() == reflect.Ptr {
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

func (m Model) Create(data any) (id string, err error) {
	m.CheckOID()
	if data != nil {
		m.Data = data
	}
	bsonData := convertToBSONM(m.Data)
	log.Debugf("创建MongoDB数据: %+v\n", bsonData)
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
func (m Model) Save(data any, value ...any) (id string, err error) {
	m.CheckOID()
	if data != nil {
		m.Data = data
	}
	bsonData := convertToBSONM(data)

	log.Debugf("MongoDB保存Where条件: %+v\n", m.WhereList)
	update := bson.M{"$set": bsonData}
	if len(value) > 0 {
		for _, data := range value {
			bm, ok := data.(bson.M)
			if ok {
				for k, v := range bm {
					update[k] = v
				}
				continue
			}
			bd, ok := data.(bson.D)
			if ok {
				for _, v := range bd {
					update[v.Key] = v.Value
				}
			}
		}
	}
	log.Debugf("MongoDB保存条件: %+v\n", update)

	opts := options.Update().SetUpsert(true)
	result, err := m.Tx.Client.Database(m.Tx.Database).Collection(m.GetCollection(m.Data)).UpdateOne(m.GetContext(), m.WhereList, update, opts)
	if err != nil {
		log.Error(err)
		return "", err
	}
	id = result.UpsertedID.(primitive.ObjectID).Hex()
	if id == "" {
		if m.WhereList["_id"] != nil {
			id = fmt.Sprint(m.WhereList["_id"])
		}
	}
	setIDField(m.Data, id)
	return
}

// 删除
func (m Model) Delete(data any) error {
	m.CheckOID()
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
func (m Model) Update(data any, value ...any) error {
	m.CheckOID()
	if data != nil {
		m.Data = data
	}
	bsonData := convertToBSONM(data)
	opts := options.Update().SetUpsert(false)
	log.Debugf("MongoDB更新Where条件: %v\n", m.WhereList)
	delete(bsonData, "_id")

	update := bson.M{"$set": bsonData}
	if len(value) > 0 {
		for _, data := range value {
			bm, ok := data.(bson.M)
			if ok {
				for k, v := range bm {
					update[k] = v
				}
				continue
			}
			bd, ok := data.(bson.D)
			if ok {
				for _, v := range bd {
					update[v.Key] = v.Value
				}
			}
		}
	}
	log.Debugf("MongoDB保存条件: %+v\n", update)

	_, err := m.Tx.Client.Database(m.Tx.Database).Collection(m.GetCollection(m.Data)).UpdateMany(m.GetContext(), m.WhereList, update, opts)

	if err != nil {
		log.Error(err)
	}
	return err
}

// 查询数据
func (m Model) Find() orm.ORMQuary {
	m.CheckOID()
	return Quary{m: m, Where: m.WhereList}
}
