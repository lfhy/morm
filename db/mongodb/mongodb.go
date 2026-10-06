package mongodb

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lfhy/morm/conf"
	"github.com/lfhy/morm/log"
	"github.com/lfhy/morm/types"
	"golang.org/x/net/proxy"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

func Init() (types.ORM, error) {
	config := &conf.MongoDBConfig{
		Database:       conf.ReadConfigToString("mongodb", "database"),
		OptionPoolSize: strconv.Itoa(conf.ReadConfigToInt("mongodb", "option_pool_size")),
		Proxy:          conf.ReadConfigToString("mongodb", "proxy"),
		Uri:            conf.ReadConfigToString("mongodb", "uri"),
		W:              conf.ReadConfigToString("mongodb", "w"),
		ReadMode:       conf.ReadConfigToString("mongodb", "readmode"),
	}
	conn, err := initWithConfig(context.Background(), config, false)
	if err != nil {
		return nil, err
	}
	ORMConn = conn.(*DBConn)
	return conn, nil
}

// InitWithConfig creates an independent connection without using or changing
// the legacy configuration singleton. The caller owns the returned connection.
func InitWithConfig(ctx context.Context, config *conf.MongoDBConfig) (types.ORM, error) {
	return initWithConfig(ctx, config, true)
}

func initWithConfig(ctx context.Context, config *conf.MongoDBConfig, verify bool) (types.ORM, error) {
	if ctx == nil {
		return nil, errors.New("mongodb initialization requires a context")
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("initialize mongodb: %w", err)
	}
	if config == nil {
		return nil, errors.New("mongodb initialization requires a configuration")
	}
	// 设置连接uri
	opts := options.Client().ApplyURI(config.Uri)
	proxys := config.Proxy
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
	poolSize, _ := strconv.Atoi(config.OptionPoolSize)
	readMode := config.ReadMode
	W := config.W
	if W == "" {
		W = "majority"
	}
	if poolSize <= 0 {
		poolSize = 150
	}

	opts.SetMaxPoolSize(uint64(poolSize))
	opts.SetMinPoolSize(uint64(poolSize / 10))
	// Bound topology discovery and individual socket establishment without
	// imposing a default deadline on every database command. Client.Timeout
	// also caps long-running operations such as CreateIndexes, so callers that
	// need an operation deadline must provide it through context instead.
	applyClientNetworkTimeouts(opts)
	// 只读取主节点
	opts.SetReadPreference(readpref.Primary())
	// 连接mongodb
	// // 写确认
	wInt, err := strconv.Atoi(W)
	if err == nil {
		opts.SetWriteConcern(&writeconcern.WriteConcern{
			W:        wInt,
			WTimeout: 1 * time.Second,
		})
	} else {
		opts.SetWriteConcern(&writeconcern.WriteConcern{
			W:        W,
			WTimeout: 1 * time.Second,
		})
	}

	// 连接mongodb
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, mongoInitError(ctx, "connect")
	}
	cleanup := func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = client.Disconnect(cleanupCtx)
	}
	if verify {
		if err := client.Ping(ctx, readpref.Primary()); err != nil {
			cleanup()
			return nil, mongoInitError(ctx, "ping primary")
		}
	}
	if err := ctx.Err(); err != nil {
		cleanup()
		return nil, fmt.Errorf("mongodb connect: %w", err)
	}
	conn := &DBConn{Database: config.Database, Client: client, NearestClient: client}
	if readMode == "master" {
		return conn, nil
	}

	// 使用就近读取
	opts.SetReadPreference(readpref.Nearest(
		readpref.WithHedgeEnabled(true),
		readpref.WithMaxStaleness(5*time.Minute),
	))
	// 连接mongodb
	nearestClient, err := mongo.Connect(ctx, opts)
	if err != nil {
		cleanup()
		return nil, mongoInitError(ctx, "connect nearest")
	}
	conn.NearestClient = nearestClient
	if verify {
		if err := nearestClient.Ping(ctx, opts.ReadPreference); err != nil {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			_ = conn.Disconnect(cleanupCtx)
			cancel()
			return nil, mongoInitError(ctx, "ping nearest")
		}
	}
	if err := ctx.Err(); err != nil {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = conn.Disconnect(cleanupCtx)
		cancel()
		return nil, fmt.Errorf("mongodb connect nearest: %w", err)
	}

	return conn, nil
}

func applyClientNetworkTimeouts(opts *options.ClientOptions) {
	if opts.ConnectTimeout == nil {
		opts.SetConnectTimeout(30 * time.Second)
	}
	if opts.ServerSelectionTimeout == nil {
		opts.SetServerSelectionTimeout(30 * time.Second)
	}
}

// Driver errors can include the connection string, including credentials.
// Preserve context cancellation for errors.Is without exposing that string.
func mongoInitError(ctx context.Context, stage string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("mongodb %s: %w", stage, err)
	}
	return fmt.Errorf("mongodb %s failed", stage)
}

// 获取集合
func (m *Model) GetCollection(dest any) string {
	if m.Collection != "" {
		return m.Collection
	}
	m.Collection = GetTableName(dest)
	return m.Collection
}

func GetTableName(dest any) string {
	switch v := dest.(type) {
	case Table:
		return v.TableName()
	case *Table:
		return (*v).TableName()
	case string:
		return fmt.Sprint(dest)
	default:
		return reflect.TypeOf(dest).Elem().Name()
	}
}

// 启动事务做函数调用
func (m *Model) Session(transactionFunc func(session types.Session) error) error {
	session, err := m.Tx.Client.StartSession()
	if err != nil {
		return log.Error(err)
	}
	ctx := m.GetContext()
	defer session.EndSession(ctx)

	// WithTransaction may retry the callback. Give each attempt its own model,
	// bound to the driver's session context, without leaving that context on the
	// original model after the transaction ends.
	_, err = session.WithTransaction(ctx, func(sessionCtx mongo.SessionContext) (any, error) {
		txModel := &Model{
			Tx:                 m.Tx,
			Data:               m.Data,
			WhereList:          cloneMongoValue(m.WhereList).(bson.M),
			Ctx:                sessionCtx,
			Collection:         m.Collection,
			transactionSession: session,
		}
		m.OpList.Range(func(key, value any) bool {
			txModel.OpList.Store(key, cloneMongoValue(value))
			return true
		})
		return nil, transactionFunc(&SessionModel{session: session, Model: txModel})
	})
	if err != nil {
		return log.Error(err)
	}
	return nil
}

type SessionModel struct {
	session mongo.Session
	*Model
}

// Clone filters/options for each WithTransaction attempt. Callback operations
// may modify nested BSON values; a retry must begin with the caller's original
// conditions, not the previous attempt's mutations.
func cloneMongoValue(value any) any {
	switch v := value.(type) {
	case bson.M:
		copy := make(bson.M, len(v))
		for key, item := range v {
			copy[key] = cloneMongoValue(item)
		}
		return copy
	case bson.D:
		copy := make(bson.D, len(v))
		for i, item := range v {
			copy[i] = bson.E{Key: item.Key, Value: cloneMongoValue(item.Value)}
		}
		return copy
	case bson.A:
		copy := make(bson.A, len(v))
		for i, item := range v {
			copy[i] = cloneMongoValue(item)
		}
		return copy
	case map[string]any:
		copy := make(map[string]any, len(v))
		for key, item := range v {
			copy[key] = cloneMongoValue(item)
		}
		return copy
	case []any:
		copy := make([]any, len(v))
		for i, item := range v {
			copy[i] = cloneMongoValue(item)
		}
		return copy
	default:
		return value
	}
}

// SwitchModel 返回绑定到当前 session 的新 ORMModel，允许跨集合操作。
func (s *SessionModel) SwitchModel(data any) types.ORMModel {
	m := &Model{
		Data:               data,
		Tx:                 s.Tx,
		WhereList:          bson.M{},
		OpList:             sync.Map{},
		Ctx:                s.GetContext(),
		Collection:         "",
		transactionSession: s.session,
	}
	m.Collection = m.GetCollection(data)
	return m
}

func (m *SessionModel) Commit() error {
	// Commit/Rollback are terminal for this callback. WithTransaction detects
	// the final state and must not issue a second commit or abort.
	return m.session.CommitTransaction(m.GetContext())
}

func (m *SessionModel) Rollback() error {
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

		// time.Time 直接写入，避免被递归转换成空对象
		if field.Type() == reflect.TypeOf(time.Time{}) {
			bsonData[fieldName] = field.Interface()
			continue
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

		if fieldName == "_id" {
			id := field.Interface()
			switch id := id.(type) {
			case primitive.ObjectID:
				// OID则不需要改
				bsonData[fieldName] = id
			default:
				// 判断是否是0值 是则跳过
				if field.IsZero() {
					continue
				}
				idstr := fmt.Sprint(id)
				oid, err := primitive.ObjectIDFromHex(idstr)
				if err != nil {
					// _id 允许显式使用字符串等自定义值；仅在看起来像 hex ObjectID
					// 时转换，失败则按原值写入。
					bsonData[fieldName] = id
					continue
				}
				bsonData[fieldName] = oid
			}
		} else {
			// 常规字段赋值
			bsonData[fieldName] = field.Interface()
		}

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

func (m *Model) Create(data any) (id string, err error) {
	m.CheckOID()
	if data != nil {
		m.Data = data
	}
	bsonData, err := ConvertToBSONM(m.Data)
	if err != nil {
		return "", err
	}
	log.Debugf("创建MongoDB数据: %+v\n", bsonData)
	result, err := m.Tx.Client.Database(m.Tx.Database).Collection(m.GetCollection(m.Data)).InsertOne(m.GetContext(), bsonData)
	if err != nil {
		log.Error(err)
		return "", err
	}
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
	_, err := m.Create(data)
	return err
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
	if len(data) > 0 {
		m.Where(data[0])
	}
	return m.Find().Delete()
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
func (m *Model) Find() types.ORMQuery {
	m.CheckOID()
	return &Query{m: m, Where: m.WhereList}
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
	_, err := m.Tx.Client.Database(m.Tx.Database).Collection(m.GetCollection(m.Data)).BulkWrite(m.GetContext(), models, bulkWriteOpts)
	return err
}
