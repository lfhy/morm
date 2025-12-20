package morm

import (
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
	model := base.M()
	switch w := any(where).(type) {
	case func(m Model):
		w(model)
	default:
		if w != nil {
			model.Where(w)
		}
	}
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

func buildWhere[Where any | func(m Model)](model Model, where Where) {
	switch f := any(where).(type) {
	case func(m Model):
		f(model)
	default:
		if f != nil {
			model.Where(f)
		}
	}
}

// 获取单个
// Where 可以是函数，也可以是Model
func One[T any](baseModel BaseModel, where ...any) (*T, error) {
	var base T
	model := baseModel.M()
	for _, fn := range where {
		buildWhere(model, fn)
	}
	return &base, model.Find().One(&base)
}

// 获取多个
// Where 可以是函数，也可以是Model
func All[T any](baseModel BaseModel, where ...any) ([]*T, error) {
	var base []*T
	model := baseModel.M()
	for _, fn := range where {
		buildWhere(model, fn)
	}
	return base, model.Find().All(&base)
}

// 删除
// Where 可以是函数，也可以是Model
func Delete(baseModel BaseModel, where any) error {
	model := baseModel.M()
	buildWhere(model, where)
	return model.Delete()
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
	model := baseModel.M()
	buildWhere(model, where)
	return model.Update(update)
}

// 更新或插入
// Where 可以是函数，也可以是Model
// update 为Model对象
func Upsert(baseModel BaseModel, where any, update any) error {
	model := baseModel.M()
	buildWhere(model, where)
	return model.Upsert(update)
}

// 创建并返回ID
func Insert(baseModel BaseModel) (id string, err error) {
	data := types.DeepCopy(baseModel)
	return baseModel.M().Create(data)
}
