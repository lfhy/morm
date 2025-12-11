package types

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

// 为了避免识别错误设置Bool类型为int类型
type BoolORM int

const (
	BoolORMTrue  BoolORM = 1
	BoolORMFalse BoolORM = -1
	// 空字符串
	EmptyStr = "-"
)

type ORM interface {
	// Model用于返回Orm模型 用于之后对模型的操作
	Model(data any) ORMModel
}

type Session interface {
	ORMModel
	Commit() error
	Rollback() error
}

type Cursor interface {
	Next() bool
	Decode(v any) error
	Close() error
}

type ORMModel interface {
	// 插入数据
	// 返回错误
	// 传入的必须是结构体指针才可以修改原始数据
	Create(data any) (string, error)

	// 插入数据
	// 等同Create
	Insert(data any) error

	// 更新或插入数据
	// 返回错误
	// 传入的必须是结构体指针才可以修改原始数据
	// 输入Where(&User{Name:"test"}).Save(&User{Value:"123"})
	// 当User表中的Name有test数据时，则修改该数据的Value为 123
	// 当User表中的Name没有test数据时，则插入该数据 User{Name:"test",Value:"123"}
	// 支持传入Save(map[string]any{"name":"123"}) 进行构建
	Save(data any, value ...any) error

	// 更新或插入数据
	// 等同Save
	Upsert(data any, value ...any) error

	// 更新
	// 返回错误 考虑更新可能更新多条 不返回ID
	// 传入的必须是结构体指针才可以修改原始数据
	// Update(&User{ID:123}) 会生成 UPDATE User SET ID = 123
	// Update("ID",123) 也会生成 UPDATE User SET ID = 123
	// Update(map[string]any{"ID":"123"}) 也会生成 UPDATE User SET ID = 123
	Update(data any, value ...any) error

	// 批量写入
	// 在mongo中datas为[]MongoBulkWriteOperation
	// 在sql中datas为[]BulkWriteOperation
	// order为写入是否是有序 mongo中使用无序可以提高性能
	BulkWrite(datas any, order bool) error

	// 事务
	// 事务中要使用sessionModel 进行操作 返回error不为 nil 时则会进行回滚
	Session(transactionFunc func(sessionModel Session) error) error

	// 上下文
	GetContext() context.Context
	SetContext(ctx context.Context) ORMModel

	// 删除
	// 传入Delete(&User{ID:123}) 就是删除ID为123的数据
	// 也可以直接Where(&User{ID:123}).Delete()
	Delete(data ...any) error

	// 查询数据
	// 会根据限制条件生成查询函数
	// 具体查询执行需要在查询函数中进行
	Find() ORMQuery

	// 清除过滤条件
	Reset() ORMModel
	ResetFilter() ORMModel

	// 过滤条件
	// Where只能传入结构体
	// 会根据每个结构体的赋值情况进行查询
	// Where(&User{ID:123}) 会生成 WHERE User.ID = 123
	// Where("ID",123) 也会生成 WHERE User.ID = 123
	// Where(map[string]any{"ID":"123"}) 也会生成 WHERE User.ID = 123
	Where(condition any, value ...any) ORMModel

	// Equal等同Where
	Equal(key any, value ...any) ORMModel

	// WhereIs传入 key和value 根据生成表达式
	// WhereIs允许用户直接操作gorm或者mongo的opList
	// 如果数据库是gorm的则可以传入gorm的语法如 WhereIs("where user = ?","123")
	// 如果数据库是mongo的则可以传入mongo的语法如 WhereIs("user",bson.M{"$eq":123})
	// 在非必要的情况下不建议使用WhereIs
	WhereIs(key string, value any) ORMModel

	// WhereNot只能传入结构体
	// 会根据每个结构体的赋值情况进行查询
	// WhereNot(&User{ID:123}) 会生成 WHERE User.ID <> 123
	// WhereNot("ID",123) 也会生成 WHERE User.ID <> 123
	// WhereNot(map[string]any{"ID":"123"}) 也会生成 WHERE User.ID <> 123
	WhereNot(condition any, value ...any) ORMModel

	// Not等同WhereNot
	Not(condition any, value ...any) ORMModel

	// WhereGt只能传入结构体
	// 会根据每个结构体的赋值情况进行查询
	// WhereGt(&User{ID:123}) 会生成 WHERE User.ID > 123
	// WhereGt("ID",123) 也会生成 WHERE User.ID > 123
	// WhereGt(map[string]any{"ID":"123"}) 也会生成 WHERE User.ID > 123
	WhereGt(condition any, value ...any) ORMModel

	// Gt等同WhereGt
	Gt(condition any, value ...any) ORMModel

	// WhereLt只能传入结构体
	// 会根据每个结构体的赋值情况进行查询
	// WhereLt(&User{ID:123}) 会生成 WHERE User.ID < 123
	// WhereLt("ID",123) 也会生成 WHERE User.ID < 123
	// WhereLt(map[string]any{"ID":"123"}) 也会生成 WHERE User.ID < 123
	WhereLt(condition any, value ...any) ORMModel

	// WhereLt等同WhereLt
	Lt(condition any, value ...any) ORMModel

	// WhereGte只能传入结构体
	// 会根据每个结构体的赋值情况进行查询
	// WhereGte(&User{ID:123}) 会生成 WHERE User.ID >= 123
	// WhereGte("ID",123) 也会生成 WHERE User.ID >= 123
	// WhereGte(map[string]any{"ID":"123"}) 也会生成 WHERE User.ID >= 123
	WhereGte(condition any, value ...any) ORMModel

	// Gte等同WhereGte
	Gte(condition any, value ...any) ORMModel

	// WhereLte只能传入结构体
	// 会根据每个结构体的赋值情况进行查询
	// WhereLte(&User{ID:123}) 会生成 WHERE User.ID <= 123
	// WhereLte("ID",123) 也会生成 WHERE User.ID <= 123
	// WhereLte(map[string]any{"ID":"123"}) 也会生成 WHERE User.ID <= 123
	WhereLte(condition any, value ...any) ORMModel

	// Lte等同WhereLte
	Lte(condition any, value ...any) ORMModel

	// WhereOr只能传入结构体
	// 会根据每个结构体的赋值情况进行查询
	// Where(&User{ID:12}).WhereOr(&User{ID:123}) 会生成 WHERE User.ID = 12 OR User.ID = 123
	// Where("ID",12).WhereOr("ID",123) 也会生成 WHERE User.ID = 12 OR User.ID = 123
	// Where(map[string]any{"ID":"12"}).WhereOr(map[string]any{"ID":"123"}) 也会生成 WHERE User.ID = 12 OR User.ID = 123
	WhereOr(condition any, value ...any) ORMModel

	// Or等同WhereOr
	Or(condition any, value ...any) ORMModel

	// 模糊查询
	// 输入WhereLike(&User{Name:"test"})
	// 会生成 WHERE User.Name LIKE "%test%"
	// 输入WhereLike("Name","_test%")
	// 会生成 WHERE User.Name LIKE "_test%"
	// Mongo则会使用$regex
	// 使用结构体默认使用primitive.Regex{Pattern: "test", Options: "i"}
	WhereLike(condition any, value ...any) ORMModel

	// 模糊查询
	// 等同WhereLike
	Like(condition any, value ...any) ORMModel

	// 限制查询的数量
	Limit(limit int) ORMModel

	// 跳过查询个数
	// 传入3 则会跳过前3个结果从第4个结果开始取
	// 需要配合Limit使用
	Offset(offset int) ORMModel

	// 正序
	// 传入结构体会根据结构体第一个有值的对象进行正序排序
	// 如传入Asc(&User{ID:1}) 就会根据User.ID进行排序 ID值不会处理但是要排序的必须有值 且只会根据第一个有值对象进行排序
	Asc(condition any) ORMModel

	// 逆序
	// 传入结构体会根据结构体第一个有值的对象进行倒序排序
	// 如传入Desc(&User{ID:1}) 就会根据User.ID进行排序 ID值不会处理但是要排序的必须有值 且只会根据第一个有值对象进行排序
	Desc(condition any) ORMModel

	// 分页
	Page(page, limit int) ORMModel

	// 查询匹配到的一条数据
	One(data any) error

	// 查询全部数据
	All(data any) error

	// 返回查询个数
	Count() int64

	// 游标
	// 在查询大量数据时可以减少内存占用
	// 使用时需要及时使用Close 避免内存泄漏
	Cursor() (Cursor, error)
}

type ORMQuery interface {
	// 查询匹配到的一条数据
	One(data any) error
	// 查询全部数据
	All(data any) error
	// 返回查询个数
	Count() int64
	// 删除查询结果
	Delete() error
	// 游标
	Cursor() (Cursor, error)
}

// 分页
func Page(m *ORMModel, page, limit int) {
	if page <= 0 {
		page = 1
	}
	(*m).Offset((page - 1) * limit).Limit(limit)
}

type BulkWriteOperation struct {
	Type   string // "insert", "update", "delete"
	Data   any
	Where  map[string]any
	Values map[string]any
}

type MongoBulkWriteOperation = mongo.WriteModel
