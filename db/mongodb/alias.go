package mongodb

import "github.com/lfhy/morm/types"

func (m *Model) Equal(key any, value ...any) types.ORMModel {
	return m.Where(key, value...)
}

func (m *Model) Like(condition any, value ...any) types.ORMModel {
	return m.WhereLike(condition, value...)
}

func (m *Model) Not(condition any, value ...any) types.ORMModel {
	return m.WhereNot(condition, value...)
}

func (m *Model) Gt(condition any, value ...any) types.ORMModel {
	return m.WhereGt(condition, value...)
}

func (m *Model) Lt(condition any, value ...any) types.ORMModel {
	return m.WhereLt(condition, value...)
}

func (m *Model) Gte(condition any, value ...any) types.ORMModel {
	return m.WhereGte(condition, value...)
}

func (m *Model) Lte(condition any, value ...any) types.ORMModel {
	return m.WhereLte(condition, value...)
}

func (m *Model) Or(condition any, value ...any) types.ORMModel {
	return m.WhereOr(condition, value...)
}

func (m *Model) Reset() types.ORMModel {
	return m.ResetFilter()
}
