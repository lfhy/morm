package morm

import (
	"errors"
	"fmt"
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
	total, err := ListWithError(base, ctx, where, listFn)
	if err != nil {
		log.Errorf("List Error:%v", err)
	}
	return total
}

// ListWithError 分页查询并返回查询、游标和回调错误。
// BSON 解码失败会跳过当前记录；回调返回错误或 false 时停止继续读取。
func ListWithError[T BaseModel, ListFn func(m T) bool | func(m T) | func(m T) error](base T, ctx *ListOption, where any, listFn ListFn) (total int64, err error) {
	model := buildWhere(base.M(), where)
	total, err = model.Find().CountWithError()
	if err != nil {
		return total, fmt.Errorf("counting list records: %w", err)
	}
	if total == 0 {
		return total, nil
	}
	if listFn == nil {
		return total, nil
	}
	if ctx == nil {
		// 空分页配置按默认值处理，避免错误感知入口因 nil 配置 panic。
		ctx = &ListOption{}
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
		return total, fmt.Errorf("opening list cursor: %w", err)
	}
	defer func() {
		if closeErr := cur.Close(); closeErr != nil {
			closeErr = fmt.Errorf("closing list cursor: %w", closeErr)
			if err == nil {
				err = closeErr
				return
			}
			err = errors.Join(err, closeErr)
		}
	}()
	for cur.Next() {
		var record T
		if decodeErr := cur.Decode(&record); decodeErr != nil {
			log.Errorf("Decode Error:%v", decodeErr)
			continue
		}
		switch lfn := any(listFn).(type) {
		case func(m T) bool:
			if !lfn(record) {
				return total, nil
			}
		case func(m T) error:
			if callbackErr := lfn(record); callbackErr != nil {
				return total, fmt.Errorf("list callback: %w", callbackErr)
			}
		case func(m T):
			lfn(record)
		}
	}
	if cursorErr := cur.Err(); cursorErr != nil {
		return total, fmt.Errorf("iterating list cursor: %w", cursorErr)
	}
	return total, nil
}

// buildWhere 支持 where 为：
//  1. func(m Model) 回调函数
//  2. Model（ORMModel 接口）：已链式构造好的查询模型，如 m.M().Lt(...).WhereIs(...)
//     此时直接复用该模型（含表名、Where 条件），后续条件继续叠加
//  3. 其他任意类型：走 model.Where(w)（结构体/map 等）
//
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
