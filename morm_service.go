package morm

import "github.com/lfhy/morm/types"

type BaseModel interface {
	M() Model
	TableName() string
}

func List[T BaseModel](base T, ctx *ListOption, where func(m Model), listFn func(m T) bool) int64 {
	model := base.M()
	if where != nil {
		where(model)
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
	cur, err := model.Cursor()
	if err != nil {
		return total
	}
	defer cur.Close()
	for cur.Next() {
		var base T
		if err := cur.Decode(&base); err != nil {
			continue
		}
		if !listFn(base) {
			break
		}
	}
	return total
}

// 获取单个
func One[T any](baseModel BaseModel, where ...func(m Model)) (*T, error) {
	var base T
	model := baseModel.M()
	for _, fn := range where {
		fn(model)
	}
	return &base, model.Find().One(&base)
}

// 获取多个
func All[T any](baseModel BaseModel, where ...func(m Model)) ([]*T, error) {
	var base []*T
	model := baseModel.M()
	for _, fn := range where {
		fn(model)
	}
	return base, model.Find().All(&base)
}

// 删除
func Delete(baseModel BaseModel, where func(m Model)) error {
	model := baseModel.M()
	if where != nil {
		where(model)
	}
	return model.Delete()
}

// 创建
func Create(baseModel BaseModel) error {
	data := types.DeepCopy(baseModel)
	_, err := baseModel.M().Create(data)
	return err
}

// 更新
func Update[T any](baseModel BaseModel, where func(m Model), update T) error {
	model := baseModel.M()
	where(model)
	return model.Update(update)
}

// 更新或插入
func Upsert[T any](baseModel BaseModel, where func(m Model), update T) error {
	model := baseModel.M()
	where(model)
	return model.Upsert(update)
}
