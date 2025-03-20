package orm

import (
	"context"
)

// 为了避免识别错误设置Bool类型为int类型
type BoolORM int

const (
	TrueInt  BoolORM = 1
	FalseInt BoolORM = -1
	// 空字符串
	EmptyStr = "-"
)

type ORM interface {
	// Model用于返回Orm模型 用于之后对模型的操作
	Model(data any) ORMModel
}

type ORMModel interface {
	// 插入数据
	// 返回ID和错误
	// 传入的必须是结构体指针才可以修改原始数据
	Create(data any) (string, error)

	// 更新或插入数据
	// 返回ID和错误
	// 传入的必须是结构体指针才可以修改原始数据
	Save(data any, value ...any) (string, error)

	// 更新
	// 返回错误 考虑更新可能更新多条 不返回ID
	// 传入的必须是结构体指针才可以修改原始数据
	Update(data any, value ...any) error

	// 事务
	Session(transactionFunc func(sessionContext context.Context) error) error

	// 上下文
	GetContext() context.Context
	SetContext(ctx context.Context) ORMModel

	// 删除
	Delete(data any) error

	// 查询数据
	// 会根据限制条件生成查询函数
	// 具体查询执行需要在查询函数中进行
	Find() ORMQuary

	// 过滤条件
	// Where只能传入结构体
	// 会根据每个结构体的赋值情况进行查询
	// 传入Where(&User{ID:123}) 会生成 WHERE User.ID = 123
	Where(condition any, value ...any) ORMModel

	// WhereIs传入 key和value 根据生成表达式
	WhereIs(key string, value any) ORMModel

	// WhereNot只能传入结构体
	// 会根据每个结构体的赋值情况进行查询
	// 传入WhereNot(&User{ID:123}) 会生成 WHERE User.ID <> 123
	WhereNot(condition any, value ...any) ORMModel

	// WhereGt只能传入结构体
	// 会根据每个结构体的赋值情况进行查询
	// 传入WhereGt(&User{ID:123}) 会生成 WHERE User.ID > 123
	WhereGt(condition any, value ...any) ORMModel

	// WhereLt只能传入结构体
	// 会根据每个结构体的赋值情况进行查询
	// 传入WhereLt(&User{ID:123}) 会生成 WHERE User.ID < 123
	WhereLt(condition any, value ...any) ORMModel

	// WhereOr只能传入结构体
	// 会根据每个结构体的赋值情况进行查询
	// 传入Where(&User{ID:12}).WhereOr(&User{ID:123}) 会生成 WHERE User.ID = 12 OR User.ID = 123
	WhereOr(condition any, value ...any) ORMModel

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
}

type ORMQuary interface {
	// 最后查询一条数据
	One(data any) error
	// 查询全部数据
	All(data any) error
	// 返回查询个数
	Count() (int64, error)
	// 删除查询结果
	Delete() error
}

// 分页
func Page(m *ORMModel, page, limit int) {
	if page <= 0 {
		page = 1
	}
	(*m).Offset((page - 1) * limit).Limit(limit)
}
