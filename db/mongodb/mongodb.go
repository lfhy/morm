package mongodb

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/lfhy/morm/conf"
	"github.com/lfhy/morm/types"
	"golang.org/x/net/proxy"

	"github.com/lfhy/morm/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

func Init() (types.ORM, error) {
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
func (m *Model) GetCollection(dest any) string {
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
func (m *Model) Session(transactionFunc func(session types.Session) error) error {
	// 创建会话
	session, err := m.Tx.Client.StartSession()
	if err != nil {
		return log.Error(err)
	}
	// 使用适配后的函数
	sctx := m.GetContext()
	defer session.EndSession(sctx)

	sessionModel := &SessionModel{session: session, Model: m}

	adaptedFunc := func(ctx mongo.SessionContext) (any, error) {
		return nil, transactionFunc(sessionModel)
	}

	_, err = session.WithTransaction(sctx, adaptedFunc)
	// 需要排除是否用户主动操作事务
	if err == nil && sessionModel.userControlTranslator {
		return nil
	}
	if err != nil {
		session.AbortTransaction(sctx)
		return log.Error(err)
	}
	session.CommitTransaction(sctx)
	return nil
}

type SessionModel struct {
	session               mongo.Session
	userControlTranslator bool
	*Model
}

func (m *SessionModel) Commit() error {
	m.userControlTranslator = true
	return m.session.CommitTransaction(m.GetContext())
}

func (m *SessionModel) Rollback() error {
	m.userControlTranslator = true
	return m.session.AbortTransaction(m.GetContext())
}

type Table interface {
	TableName() string
}

// 转换data TO bsonM
func ConvertToBSONM(data any) (bson.M, error) {
	if data == nil {
		return bson.M{}, nil
	}

	bm, ok := data.(bson.M)
	if ok {
		return bm, nil
	}
	bsonData := bson.M{}
	bd, ok := data.(bson.D)
	if ok {
		d, _ := bson.Marshal(bd)
		bson.Unmarshal(d, &bsonData)
		return bsonData, nil
	}
	val := reflect.ValueOf(data)

	// 解引用接口
	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}

	// 解引用指针直到获取实际值
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			val = reflect.New(val.Type())
		}
		val = val.Elem()
	}

	// 仅处理结构体类型
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("input must be a struct or pointer to struct")
	}

	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		structField := typ.Field(i)

		// 解析 bson 标签（处理 `bson:"fieldName,omitempty"` 格式）
		tag, ok := structField.Tag.Lookup("bson")
		if !ok {
			continue
		}
		parts := strings.Split(tag, ",")
		fieldName := parts[0]
		if fieldName == "" {
			fieldName = strings.ToLower(structField.Name) // 默认字段名
		}

		// 是否设置零值（根据 must 标志）
		must := false
		for _, part := range parts[1:] {
			if part == "must" {
				must = true
				break
			}
		}

		// 检查零值并跳过（若需要）
		if !must && isZero(field) {
			continue
		}

		// 处理指针问题
		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				field = reflect.New(field.Type())
			}
			field = field.Elem()
		}

		// 处理嵌套结构体或指针（递归转换）
		if field.Kind() == reflect.Struct {
			nestedData, err := ConvertToBSONM(field.Interface())
			if err != nil {
				return nil, err
			}
			bsonData[fieldName] = nestedData
			continue
		}

		// 常规字段赋值
		bsonData[fieldName] = field.Interface()
	}

	return bsonData, nil
}

// 辅助函数：判断零值
func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	default:
		return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
	}
}

func (m *Model) Create(data any) (err error) {
	m.CheckOID()
	if data != nil {
		m.Data = data
	}
	bsonData, err := ConvertToBSONM(m.Data)
	if err != nil {
		return err
	}
	log.Debugf("创建MongoDB数据: %+v\n", bsonData)
	result, err := m.Tx.Client.Database(m.Tx.Database).Collection(m.GetCollection(m.Data)).InsertOne(m.GetContext(), bsonData)
	if err != nil {
		log.Error(err)
		return err
	}
	var id string
	if result.InsertedID == nil {
		if m.WhereList["_id"] != nil {
			id = fmt.Sprint(m.WhereList["_id"])
		}
	} else {
		switch result.InsertedID.(type) {
		case primitive.ObjectID:
			id = result.InsertedID.(primitive.ObjectID).Hex()
		case string:
			id = result.InsertedID.(string)
		default:
			id = fmt.Sprint(result.InsertedID)
		}
	}
	setIDField(m.Data, id)
	// log.Debugf("写入后的Date数据: %+v\n", m.Data)
	return
}

func (m *Model) Insert(data any) error {
	return m.Create(data)
}

// 更新或插入数据
func (m *Model) Save(data any, value ...any) (err error) {
	m.CheckOID()
	if data != nil {
		m.Data = data
	}
	bsonData, err := ConvertToBSONM(data)
	if err != nil {
		return err
	}
	log.Debugf("MongoDB保存Where条件: %+v\n", m.WhereList)
	update := make(bson.M)
	if len(bsonData) != 0 {
		update["$set"] = bsonData
	}
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
		return err
	}
	var id string
	if result.UpsertedID == nil {
		if m.WhereList["_id"] != nil {
			id = fmt.Sprint(m.WhereList["_id"])
		}
	} else {
		switch result.UpsertedID.(type) {
		case primitive.ObjectID:
			id = result.UpsertedID.(primitive.ObjectID).Hex()
		case string:
			id = result.UpsertedID.(string)
		default:
			id = fmt.Sprint(result.UpsertedID)
		}
	}
	setIDField(m.Data, id)
	return
}

func (m *Model) Upsert(data any, value ...any) error {
	return m.Save(data, value...)
}

// 删除
func (m *Model) Delete(data ...any) error {
	m.CheckOID()
	if len(data) > 0 && data[0] != nil {
		m.Data = data[0]
	}
	_, err := m.Tx.Client.Database(m.Tx.Database).Collection(m.GetCollection(m.Data)).DeleteMany(m.GetContext(), m.WhereList)
	if err != nil {
		log.Error(err)
	}

	return err
}

// 修改
func (m *Model) Update(data any, value ...any) error {
	m.CheckOID()
	if data != nil {
		m.Data = data
	}
	bsonData, err := ConvertToBSONM(m.Data)
	if err != nil {
		return err
	}
	delete(bsonData, "_id")
	log.Debugf("MongoDB更新bsonData: %+v\n", bsonData)
	opts := options.Update().SetUpsert(false)
	log.Debugf("MongoDB更新Where条件: %v\n", m.WhereList)
	update := make(bson.M)
	if len(bsonData) != 0 {
		update["$set"] = bsonData
	}
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
	log.Debugf("MongoDB更新条件: %+v\n", update)

	if len(update) == 0 {
		log.Error("MongoDB更新条件为空")
		return nil
	}

	_, err = m.Tx.Client.Database(m.Tx.Database).Collection(m.GetCollection(m.Data)).UpdateMany(m.GetContext(), m.WhereList, update, opts)
	if err != nil {
		log.Error(err)
	}
	return err
}

// 查询数据
func (m *Model) Find() types.ORMQuary {
	m.CheckOID()
	return &Quary{m: m, Where: m.WhereList}
}

func (m *Model) One(data any) error {
	return m.Find().One(data)
}

func (m *Model) All(data any) error {
	return m.Find().All(data)
}

func (m *Model) Count() int64 {
	return m.Find().Count()
}

func (m *Model) Cursor() (types.Cursor, error) {
	return m.Find().Cursor()
}

func (m *Model) BulkWrite(datas any, order bool) error {
	models, ok := datas.([]mongo.WriteModel)
	if !ok {
		return errors.New("datas must be []mongo.WriteModel")
	}
	// 不需要写入时，直接返回
	if len(models) == 0 {
		return nil
	}
	m.CheckOID()

	// 执行批量写入操作
	bulkWriteOpts := options.BulkWrite().SetOrdered(order) // 设置为无序时 提高性能
	_, err := m.Tx.Client.Database(m.Tx.Database).Collection(m.GetCollection(m.Data)).BulkWrite(context.TODO(), models, bulkWriteOpts)
	return err
}
