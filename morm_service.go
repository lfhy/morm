package morm

import (
	"reflect"

	"github.com/lfhy/morm/log"
	"github.com/lfhy/morm/types"
)

type BaseModel interface {
	M() Model
	TableName() string
}

// List 分页查询
// Where 可以是函数，也可以是Model
func List[T BaseModel, ListFn func(m T) bool | func(m T) | func(m T) error](base T, ctx *ListOption, where any, listFn ListFn) int64 {
	model := buildWhere(base.M(), where)
	total := model.Find().Count()
	if total == 0 {
		return total
	}
	if listFn == nil {
		return total
	}
	if !ctx.All {
		model.Page(ctx.GetPage(), ctx.GetLimit())
	}

	if ctx.Sort != nil {
		if ctx.Sort.Mode == types.OrderDirDesc {
			model.Desc(ctx.Sort.Key)
		} else {
			model.Asc(ctx.Sort.Key)
		}
	}

	for _, sort := range ctx.Sorts {
		if sort.Mode == types.OrderDirDesc {
			model.Desc(sort.Key)
		} else {
			model.Asc(sort.Key)
		}
	}
	cur, err := model.Cursor()
	if err != nil {
		log.Errorf("Cursor Error:%v", err)
		return total
	}
	defer cur.Close()
	for cur.Next() {
		var base T
		if err := cur.Decode(&base); err != nil {
			log.Errorf("Decode Error:%v", err)
			continue
		}
		switch lfn := any(listFn).(type) {
		case func(m T) bool:
			if !lfn(base) {
				break
			}
		case func(m T) error:
			if err := lfn(base); err != nil {
				log.Error("ListFn:", err)
				break
			}
		case func(m T):
			lfn(base)
		}
	}
	return total
}

// buildWhere 支持 where 为：
//  1. func(m Model) 回调函数
//  2. Model（ORMModel 接口）：已链式构造好的查询模型，如 m.M().Lt(...).WhereIs(...)
//     此时直接复用该模型（含表名、Where 条件），后续条件继续叠加
//  3. 其他任意类型：走 model.Where(w)（结构体/map 等）
// where 为 nil（含接口内 nil 指针）时跳过
func buildWhere[Where any | func(m Model)](model Model, where Where) Model {
	switch f := any(where).(type) {
	case func(m Model):
		f(model)
	case Model:
		if !isNilWhere(f) {
			model = f
		}
	default:
		if !isNilWhere(f) {
			model.Where(f)
		}
	}
	return model
}

// isNilWhere 判断接口值是否为 nil（既包括接口本身为 nil，也包括接口里包着 nil 指针的情况）
// 这样 var where morm.ORMModel 声明但未赋值时，不会误传给 Where 导致空指针
func isNilWhere(w any) bool {
	if w == nil {
		return true
	}
	v := reflect.ValueOf(w)
	switch v.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func, reflect.Interface:
		return v.IsNil()
	}
	return false
}

// 获取单个
// Where 可以是函数，也可以是Model
func One[T any](baseModel BaseModel, where ...any) (*T, error) {
	var base T
	model := baseModel.M()
	for _, fn := range where {
		model = buildWhere(model, fn)
	}
	return &base, model.Find().One(&base)
}

// 获取多个
// Where 可以是函数，也可以是Model
func All[T any](baseModel BaseModel, where ...any) ([]*T, error) {
	var base []*T
	model := baseModel.M()
	for _, fn := range where {
		model = buildWhere(model, fn)
	}
	return base, model.Find().All(&base)
}

// 删除
// Where 可以是函数，也可以是Model
func Delete(baseModel BaseModel, where any) error {
	return buildWhere(baseModel.M(), where).Delete()
}

// 创建
func Create(baseModel BaseModel) error {
	data := types.DeepCopy(baseModel)
	_, err := baseModel.M().Create(data)
	if err != nil {
		log.Errorf("Create Error:%v", err)
	}
	return err
}

// 更新
// Where 可以是函数，也可以是Model
// update 为Model对象
func Update(baseModel BaseModel, where any, update any) error {
	return buildWhere(baseModel.M(), where).Update(update)
}

// 更新或插入
// Where 可以是函数，也可以是Model
// update 为Model对象
func Upsert(baseModel BaseModel, where any, update any) error {
	return buildWhere(baseModel.M(), where).Upsert(update)
}

// 创建并返回ID
func Insert(baseModel BaseModel) (id string, err error) {
	data := types.DeepCopy(baseModel)
	return baseModel.M().Create(data)
}
