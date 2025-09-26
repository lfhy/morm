package sqlorm

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/lfhy/morm/types"

	"gorm.io/gorm"
)

const (
	WhereIs = iota
	WhereNot
	WhereGt
	WhereLt
	WhereOr
	OrderAsc
	OrderDesc
	WhereGte
	WhereLte
)

// 限制条件
func (m *Model) Where(condition any, value ...any) types.ORMModel {
	if len(value) > 0 {
		key, ok := condition.(string)
		if ok {
			m.OpList.Store(fmt.Sprintf("where %s = ?", key), value[0])
			m.upsertOp.Store(fmt.Sprintf("where %s = ?", key), value[0])
			return m
		}
	}
	return m.whereMode(condition, WhereIs)
}

func (m *Model) WhereIs(key string, value any) types.ORMModel {
	m.OpList.Store(key, value)
	return m
}

func (m *Model) WhereNot(condition any, value ...any) types.ORMModel {
	if len(value) > 0 {
		key, ok := condition.(string)
		if ok {
			m.OpList.Store(fmt.Sprintf("not %s = ?", key), value[0])
			return m
		}
	}
	return m.whereMode(condition, WhereNot)
}

func (m *Model) WhereGt(condition any, value ...any) types.ORMModel {
	if len(value) > 0 {
		key, ok := condition.(string)
		if ok {
			m.OpList.Store(fmt.Sprintf("where %s > ?", key), value[0])
			return m
		}
	}
	return m.whereMode(condition, WhereGt)
}

func (m *Model) WhereLt(condition any, value ...any) types.ORMModel {
	if len(value) > 0 {
		key, ok := condition.(string)
		if ok {
			m.OpList.Store(fmt.Sprintf("where %s < ?", key), value[0])
			return m
		}
	}
	return m.whereMode(condition, WhereLt)
}

func (m *Model) WhereGte(condition any, value ...any) types.ORMModel {
	if len(value) > 0 {
		key, ok := condition.(string)
		if ok {
			m.OpList.Store(fmt.Sprintf("where %s >= ?", key), value[0])
			return m
		}
	}
	return m.whereMode(condition, WhereGte)
}

func (m *Model) WhereLte(condition any, value ...any) types.ORMModel {
	if len(value) > 0 {
		key, ok := condition.(string)
		if ok {
			m.OpList.Store(fmt.Sprintf("where %s <= ?", key), value[0])
			return m
		}
	}
	return m.whereMode(condition, WhereLte)
}

func (m *Model) WhereOr(condition any, value ...any) types.ORMModel {
	if len(value) > 0 {
		key, ok := condition.(string)
		if ok {
			m.OpList.Store(fmt.Sprintf("or %s = ?", key), value[0])
			return m
		}
	}
	return m.whereMode(condition, WhereOr)
}

// 限制查询的数量
func (m *Model) Limit(limit int) types.ORMModel {
	m.OpList.Store("limit ", limit)
	return m
}

// 跳过查询的数量
func (m *Model) Offset(offset int) types.ORMModel {
	m.OpList.Store("offset ", offset)
	return m
}

// 正序
func (m *Model) Asc(condition any) types.ORMModel {
	return m.whereMode(condition, OrderAsc)
}

// 逆序
func (m *Model) Desc(condition any) types.ORMModel {
	return m.whereMode(condition, OrderDesc)
}
func (m *Model) whereMode(condition any, mode int) types.ORMModel {
	t := reflect.ValueOf(condition)
	if t.Kind() == reflect.Pointer {
		if t.IsNil() {
			t = reflect.New(t.Type())
		}
		t = t.Elem()
	}
	if t.Kind() == reflect.Slice {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Struct:
	l1:
		for i := 0; i < t.NumField(); i++ {
			if t.Field(i).IsZero() {
				continue
			}
			dtype := t.Type()
			value := dtype.Field(i)
			v, ok := value.Tag.Lookup("gorm")
			if ok {
				for _, v2 := range strings.Split(v, ";") {
					if strings.HasPrefix(v2, "column:") {
						switch mode {
						case WhereIs:
							column := strings.TrimPrefix(v2, "column:")
							value := t.Field(i).Interface()
							m.upsertOp.Store(column, value)
							m.OpList.Store(fmt.Sprintf("where %s = ?", column), value)
						case WhereNot:
							m.OpList.Store(fmt.Sprintf("not %s = ?", strings.TrimPrefix(v2, "column:")), t.Field(i).Interface())
						case WhereGt:
							m.OpList.Store(fmt.Sprintf("where %s > ?", strings.TrimPrefix(v2, "column:")), t.Field(i).Interface())
						case WhereLt:
							m.OpList.Store(fmt.Sprintf("where %s < ?", strings.TrimPrefix(v2, "column:")), t.Field(i).Interface())
						case WhereOr:
							m.OpList.Store(fmt.Sprintf("or %s = ?", strings.TrimPrefix(v2, "column:")), t.Field(i).Interface())
						case OrderAsc:
							m.OpList.Store(fmt.Sprintf("asc %s", strings.TrimPrefix(v2, "column:")), "")
						case OrderDesc:
							m.OpList.Store(fmt.Sprintf("desc %s", strings.TrimPrefix(v2, "column:")), "")
						case WhereGte:
							m.OpList.Store(fmt.Sprintf("where %s >= ?", strings.TrimPrefix(v2, "column:")), t.Field(i).Interface())
						case WhereLte:
							m.OpList.Store(fmt.Sprintf("where %s <= ?", strings.TrimPrefix(v2, "column:")), t.Field(i).Interface())
						}
						continue l1
					} else if !strings.Contains(v2, ":") {
						// 直接输入字段信息的情况
						switch mode {
						case WhereIs:
							column := v2
							value := t.Field(i).Interface()
							m.upsertOp.Store(column, value)
							m.OpList.Store(fmt.Sprintf("where %s = ?", column), value)
						case WhereNot:
							m.OpList.Store(fmt.Sprintf("not %s = ?", v2), t.Field(i).Interface())
						case WhereGt:
							m.OpList.Store(fmt.Sprintf("where %s > ?", v2), t.Field(i).Interface())
						case WhereLt:
							m.OpList.Store(fmt.Sprintf("where %s < ?", v2), t.Field(i).Interface())
						case WhereOr:
							m.OpList.Store(fmt.Sprintf("or %s = ?", v2), t.Field(i).Interface())
						case OrderAsc:
							m.OpList.Store(fmt.Sprintf("asc %s", v2), "")
						case OrderDesc:
							m.OpList.Store(fmt.Sprintf("desc %s", v2), "")
						case WhereGte:
							m.OpList.Store(fmt.Sprintf("where %s >= ?", v2), t.Field(i).Interface())
						case WhereLte:
							m.OpList.Store(fmt.Sprintf("where %s <= ?", v2), t.Field(i).Interface())
						}
						continue l1

					}
				}
			}
		}
	}
	return m
}

// 自动生成查询条件
func (m *Model) makeQuary() *gorm.DB {
	quary := m.getDB().Model(m.Data)
	m.OpList.Range(func(key, value any) bool {
		if strings.HasPrefix(key.(string), "where ") {
			quary = quary.Where(strings.TrimPrefix(key.(string), "where "), value)
			return true
		}
		if strings.HasPrefix(key.(string), "or ") {
			quary = quary.Or(strings.TrimPrefix(key.(string), "or "), value)
			return true
		}
		if strings.HasPrefix(key.(string), "not ") {
			quary = quary.Not(strings.TrimPrefix(key.(string), "not "), value)
			return true
		}
		if strings.HasPrefix(key.(string), "limit ") {
			quary = quary.Limit(value.(int))
			return true
		}
		if strings.HasPrefix(key.(string), "offset ") {
			quary = quary.Offset(value.(int))
			return true
		}
		if strings.HasPrefix(key.(string), "asc ") {
			quary = quary.Order(fmt.Sprintf("%s ASC", strings.TrimPrefix(key.(string), "asc ")))
			return true
		}
		if strings.HasPrefix(key.(string), "desc ") {
			quary = quary.Order(fmt.Sprintf("%s DESC", strings.TrimPrefix(key.(string), "desc ")))
			return true
		}
		// fmt.Println(key, value)
		return true
	})
	return quary
}

func (m *Model) getID(condition any) (id string) {
	t := reflect.ValueOf(condition)
	if t.Kind() == reflect.Pointer {
		if t.IsNil() {
			t = reflect.New(t.Type())
		}
		t = t.Elem()
	}
	if t.Kind() == reflect.Slice {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			if t.Field(i).IsZero() {
				continue
			}
			dtype := t.Type()
			value := dtype.Field(i)
			v, ok := value.Tag.Lookup("gorm")
			if ok {
				for _, v2 := range strings.Split(v, ";") {
					if strings.HasPrefix(v2, "id") || strings.HasPrefix(v2, "column:id") || strings.HasPrefix(v2, "primaryKey") {
						return fmt.Sprint(t.Field(i).Interface())
					}
				}
			}
		}
	}
	return
}
