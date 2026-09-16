package test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/lfhy/morm"
	"github.com/lfhy/morm/db/mongodb"
	"github.com/lfhy/morm/db/sqlorm"
	"github.com/lfhy/morm/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// listTestRecord 是错误感知列表测试使用的最小记录模型。
type listTestRecord struct {
	model morm.Model
	Value int
}

func (r listTestRecord) M() morm.Model {
	return r.model
}

func (listTestRecord) TableName() string {
	return "list_test_records"
}

// listTestModel 只替换列表所需的查询操作，其余 ORM 行为由嵌入接口提供。
type listTestModel struct {
	morm.ORMModel
	query      morm.ORMQuery
	cursor     types.Cursor
	cursorErr  error
	cursorCall int
}

func (m *listTestModel) Page(int, int) morm.ORMModel {
	return m
}

func (m *listTestModel) Asc(any) morm.ORMModel {
	return m
}

func (m *listTestModel) Desc(any) morm.ORMModel {
	return m
}

func (m *listTestModel) Find() morm.ORMQuery {
	return m.query
}

func (m *listTestModel) Cursor() (types.Cursor, error) {
	m.cursorCall++
	return m.cursor, m.cursorErr
}

type listTestQuery struct {
	morm.ORMQuery
	total int64
	err   error
}

func (q *listTestQuery) CountWithError() (int64, error) {
	return q.total, q.err
}

type listTestCursor struct {
	records     []listTestRecord
	decodeError map[int]error
	index       int
	closeErr    error
	cursorErr   error
	closed      bool
}

func (c *listTestCursor) Next() bool {
	return c.index < len(c.records)
}

func (c *listTestCursor) Decode(v any) error {
	index := c.index
	c.index++
	if err := c.decodeError[index]; err != nil {
		return err
	}

	record, ok := v.(*listTestRecord)
	if !ok {
		return fmt.Errorf("unexpected decode target %T", v)
	}
	*record = c.records[index]
	return nil
}

func (c *listTestCursor) Close() error {
	c.closed = true
	return c.closeErr
}

func (c *listTestCursor) Err() error {
	return c.cursorErr
}

func newListTestBase(total int64, cursor types.Cursor, cursorErr error) (listTestRecord, *listTestModel) {
	query := &listTestQuery{total: total}
	model := &listTestModel{
		query:     query,
		cursor:    cursor,
		cursorErr: cursorErr,
	}
	return listTestRecord{model: model}, model
}

func TestListWithErrorPropagatesCountError(t *testing.T) {
	wantErr := errors.New("count failed")
	base, model := newListTestBase(7, nil, nil)
	model.query = &listTestQuery{total: 7, err: wantErr}

	total, err := morm.ListWithError(base, nil, nil, func(listTestRecord) {})
	if total != 7 {
		t.Fatalf("total = %d, want 7", total)
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want wrapped count error", err)
	}
	if model.cursorCall != 0 {
		t.Fatalf("cursor calls = %d, want 0 after count failure", model.cursorCall)
	}
}

func TestListWithErrorPropagatesCursorErrors(t *testing.T) {
	openErr := errors.New("open cursor failed")
	iterateErr := errors.New("iterate cursor failed")
	closeErr := errors.New("close cursor failed")
	tests := []struct {
		name      string
		openErr   error
		cursorErr error
		closeErr  error
		wantErr   error
	}{
		{
			name:    "open cursor",
			openErr: openErr,
			wantErr: openErr,
		},
		{
			name:      "iterate cursor",
			cursorErr: iterateErr,
			wantErr:   iterateErr,
		},
		{
			name:     "close cursor",
			closeErr: closeErr,
			wantErr:  closeErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursor := &listTestCursor{
				closeErr:  tt.closeErr,
				cursorErr: tt.cursorErr,
			}
			base, model := newListTestBase(1, cursor, nil)
			if tt.openErr != nil {
				model.cursorErr = tt.openErr
				cursor.cursorErr = nil
			}

			total, err := morm.ListWithError(base, &morm.ListOption{All: true}, nil, func(listTestRecord) {})
			if total != 1 {
				t.Fatalf("total = %d, want 1", total)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if !cursor.closed && tt.openErr == nil {
				t.Fatal("cursor was not closed")
			}
		})
	}
}

func TestListWithErrorPropagatesCallbackErrorAndStops(t *testing.T) {
	wantErr := errors.New("callback failed")
	cursor := &listTestCursor{
		records: []listTestRecord{{Value: 1}, {Value: 2}},
	}
	base, _ := newListTestBase(2, cursor, nil)
	calls := 0

	total, err := morm.ListWithError(base, &morm.ListOption{All: true}, nil, func(listTestRecord) error {
		calls++
		return wantErr
	})
	if total != 2 {
		t.Fatalf("total = %d, want 2", total)
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want wrapped callback error", err)
	}
	if calls != 1 || cursor.index != 1 {
		t.Fatalf("callback calls = %d, cursor index = %d, want 1 and 1", calls, cursor.index)
	}
	if !cursor.closed {
		t.Fatal("cursor was not closed after callback failure")
	}
}

func TestListWithErrorStopsOnFalseCallback(t *testing.T) {
	cursor := &listTestCursor{
		records: []listTestRecord{{Value: 1}, {Value: 2}},
	}
	base, _ := newListTestBase(2, cursor, nil)
	calls := 0

	total, err := morm.ListWithError(base, &morm.ListOption{All: true}, nil, func(listTestRecord) bool {
		calls++
		return false
	})
	if err != nil {
		t.Fatalf("error = %v, want nil", err)
	}
	if total != 2 || calls != 1 || cursor.index != 1 {
		t.Fatalf("total = %d, calls = %d, cursor index = %d, want 2, 1, 1", total, calls, cursor.index)
	}
}

func TestListWithErrorSkipsDecodeErrors(t *testing.T) {
	// 解码失败路径会记录日志，测试先安装静默 logger，避免依赖生产配置。
	morm.SetDBLoger(logger.Discard)
	decodeErr := errors.New("decode failed")
	cursor := &listTestCursor{
		records:     []listTestRecord{{Value: 1}, {Value: 2}, {Value: 3}},
		decodeError: map[int]error{1: decodeErr},
	}
	base, _ := newListTestBase(3, cursor, nil)
	var values []int

	total, err := morm.ListWithError(base, &morm.ListOption{All: true}, nil, func(record listTestRecord) {
		values = append(values, record.Value)
	})
	if err != nil {
		t.Fatalf("error = %v, want nil", err)
	}
	if total != 3 {
		t.Fatalf("total = %d, want 3", total)
	}
	if want := []int{1, 3}; !reflect.DeepEqual(values, want) {
		t.Fatalf("values = %v, want %v", values, want)
	}
}

type missingListTable struct {
	ID int
}

func (missingListTable) TableName() string {
	return "missing_list_table"
}

func TestSQLCountWithErrorReturnsDatabaseError(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:count_with_error?mode=memory&cache=shared"), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sqlite connection: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	conn := &sqlorm.DBConn{DB: db}

	_, err = conn.Model(&missingListTable{}).Find().CountWithError()
	if err == nil {
		t.Fatal("CountWithError returned nil for a missing table")
	}
}

func TestMongoCountWithErrorReturnsDatabaseError(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock).
		ClientOptions(options.Client().SetRetryReads(false).SetRetryWrites(false)))
	mt.RunOpts("count error", mtest.NewOptions().DatabaseName("count_with_error").CollectionName("missing_list_table"), func(mt *mtest.T) {
		morm.SetDBLoger(logger.Discard)
		conn := &mongodb.DBConn{
			Database:      mt.DB.Name(),
			Client:        mt.Client,
			NearestClient: mt.Client,
		}
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    123,
			Message: "count failed",
		}))

		_, err := conn.Model(&missingListTable{}).Find().CountWithError()
		if err == nil {
			t.Fatal("CountWithError returned nil for a Mongo command error")
		}
	})
}

var mongoListDB morm.ORM

// mongoListRecord 验证 Mongo 列表的 Count、Cursor 和坏 BSON 处理顺序。
type mongoListRecord struct {
	Value int `bson:"value"`
}

func (mongoListRecord) TableName() string {
	return "mongo_list_records"
}

func (mongoListRecord) M() morm.Model {
	return mongoListDB.Model(&mongoListRecord{})
}

func TestMongoListWithErrorUsesCountAndCursor(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock).
		ClientOptions(options.Client().SetRetryReads(false).SetRetryWrites(false)))
	mt.RunOpts("list records", mtest.NewOptions().DatabaseName("list_with_error").CollectionName("mongo_list_records"), func(mt *mtest.T) {
		morm.SetDBLoger(logger.Discard)
		mongoListDB = &mongodb.DBConn{
			Database:      mt.DB.Name(),
			Client:        mt.Client,
			NearestClient: mt.Client,
		}
		mt.Cleanup(func() {
			mongoListDB = nil
		})

		namespace := mt.DB.Name() + ".mongo_list_records"
		mt.AddMockResponses(
			mtest.CreateCursorResponse(
				0,
				namespace,
				mtest.FirstBatch,
				bson.D{{Key: "n", Value: int64(2)}},
			),
			mtest.CreateCursorResponse(
				0,
				namespace,
				mtest.FirstBatch,
				bson.D{{Key: "value", Value: "bad"}},
				bson.D{{Key: "value", Value: 7}},
			),
		)

		var values []int
		total, err := morm.ListWithError(
			mongoListRecord{},
			&morm.ListOption{All: true},
			nil,
			func(record mongoListRecord) {
				values = append(values, record.Value)
			},
		)
		if err != nil {
			t.Fatalf("ListWithError error = %v", err)
		}
		if total != 2 {
			t.Fatalf("total = %d, want 2", total)
		}
		if want := []int{7}; !reflect.DeepEqual(values, want) {
			t.Fatalf("values = %v, want %v", values, want)
		}
	})
}

var sqlListDB *sqlorm.DBConn

// sqlListRecord 验证错误感知列表仍沿用原有分页和单字段排序。
type sqlListRecord struct {
	ID   int    `gorm:"column:id;primaryKey;autoIncrement"`
	Name string `gorm:"column:name"`
}

func (sqlListRecord) TableName() string {
	return "sql_list_records"
}

func (sqlListRecord) M() morm.Model {
	return sqlListDB.Model(&sqlListRecord{})
}

func TestListWithErrorUsesSQLPaginationAndSorting(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:list_with_error?mode=memory&cache=shared"), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sqlite connection: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	if err := db.AutoMigrate(&sqlListRecord{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	conn := &sqlorm.DBConn{DB: db}
	sqlListDB = conn
	t.Cleanup(func() {
		sqlListDB = nil
	})
	for _, name := range []string{"d", "a", "c", "b", "e"} {
		if _, err := conn.Model(&sqlListRecord{}).Create(&sqlListRecord{Name: name}); err != nil {
			t.Fatalf("create %q: %v", name, err)
		}
	}

	var names []string
	total, err := morm.ListWithError(
		sqlListRecord{},
		&morm.ListOption{
			Page:  2,
			Limit: 2,
			Sort:  &morm.Sort{Key: "name", Mode: morm.Asc},
		},
		nil,
		func(record sqlListRecord) {
			names = append(names, record.Name)
		},
	)
	if err != nil {
		t.Fatalf("ListWithError error = %v", err)
	}
	if total != 5 {
		t.Fatalf("total = %d, want 5", total)
	}
	if want := []string{"c", "d"}; !reflect.DeepEqual(names, want) {
		t.Fatalf("names = %v, want %v", names, want)
	}
}
