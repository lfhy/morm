package sqlorm

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/lfhy/morm/types"

	"gorm.io/gorm"
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
	return m.whereMode(condition, types.WhereIs)
}

func (m *Model) WhereIs(key string, value any) types.ORMModel {
	m.OpList.Store(key, value)
	return m
}

func (m *Model) WhereLike(condition any, value ...any) types.ORMModel {
	if len(value) > 0 {
		key, ok := condition.(string)
		if ok {
			m.OpList.Store(fmt.Sprintf("where %s like ?", key), value[0])
			return m
		}
	}
	return m.whereMode(condition, types.WhereLike)
}

func (m *Model) WhereNot(condition any, value ...any) types.ORMModel {
	if len(value) > 0 {
		key, ok := condition.(string)
		if ok {
			m.OpList.Store(fmt.Sprintf("not %s = ?", key), value[0])
			return m
		}
	}
	return m.whereMode(condition, types.WhereNot)
}

func (m *Model) WhereGt(condition any, value ...any) types.ORMModel {
	if len(value) > 0 {
		key, ok := condition.(string)
		if ok {
			m.OpList.Store(fmt.Sprintf("where %s > ?", key), value[0])
			return m
		}
	}
	return m.whereMode(condition, types.WhereGt)
}

func (m *Model) WhereLt(condition any, value ...any) types.ORMModel {
	if len(value) > 0 {
		key, ok := condition.(string)
		if ok {
			m.OpList.Store(fmt.Sprintf("where %s < ?", key), value[0])
			return m
		}
	}
	return m.whereMode(condition, types.WhereLt)
}

func (m *Model) WhereGte(condition any, value ...any) types.ORMModel {
	if len(value) > 0 {
		key, ok := condition.(string)
		if ok {
			m.OpList.Store(fmt.Sprintf("where %s >= ?", key), value[0])
			return m
		}
	}
	return m.whereMode(condition, types.WhereGte)
}

func (m *Model) WhereLte(condition any, value ...any) types.ORMModel {
	if len(value) > 0 {
		key, ok := condition.(string)
		if ok {
			m.OpList.Store(fmt.Sprintf("where %s <= ?", key), value[0])
			return m
		}
	}
	return m.whereMode(condition, types.WhereLte)
}

func (m *Model) WhereOr(condition any, value ...any) types.ORMModel {
	if len(value) > 0 {
		key, ok := condition.(string)
		if ok {
			m.OpList.Store(fmt.Sprintf("or %s = ?", key), value[0])
			return m
		}
	}
	return m.whereMode(condition, types.WhereOr)
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
	key, ok := condition.(string)
	if ok {
		m.OpList.Store(fmt.Sprintf("asc %s", key), "")
		return m
	}
	return m.whereMode(condition, types.OrderAsc)
}

// 逆序
func (m *Model) Desc(condition any) types.ORMModel {
	key, ok := condition.(string)
	if ok {
		m.OpList.Store(fmt.Sprintf("desc %s", key), "")
		return m
	}
	return m.whereMode(condition, types.OrderDesc)
}
func (m *Model) whereMode(condition any, mode types.WhereMode) types.ORMModel {
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
		// 处理结构体的字段，包括嵌套的匿名结构体
		var processStructFields func(reflect.Value, reflect.Type)
		processStructFields = func(val reflect.Value, typ reflect.Type) {
		l1:
			for i := 0; i < val.NumField(); i++ {
				field := val.Field(i)
				fieldType := typ.Field(i)

				// 如果是嵌套的匿名结构体，递归处理
				if field.Kind() == reflect.Struct && fieldType.Anonymous {
					processStructFields(field, fieldType.Type)
					continue
				}

				if field.IsZero() {
					continue
				}

				v, ok := fieldType.Tag.Lookup("gorm")
				if ok {
					for _, v2 := range strings.Split(v, ";") {
						if strings.HasPrefix(v2, "column:") {
							column := strings.TrimPrefix(v2, "column:")
							value := field.Interface()
							m.saveOplist(mode, column, value)
							continue l1
						} else if !strings.Contains(v2, ":") {
							// 直接输入字段信息的情况
							column := v2
							value := field.Interface()
							m.saveOplist(mode, column, value)
							continue l1
						}
					}
				}
			}
		}

		// 开始处理当前结构体的字段
		processStructFields(t, t.Type())
	case reflect.Map:
		// 遍历map
		for _, k := range t.MapKeys() {
			v := t.MapIndex(k)
			m.saveOplist(mode, k.String(), v.Interface())
		}
	}

	return m
}

// 自动生成查询条件
func (m *Model) makeQuery() *gorm.DB {
	query := m.getDB().Model(m.Data)
	m.OpList.Range(func(key string, value any) bool {
		if strings.HasPrefix(key, "where ") {
			query = query.Where(strings.TrimPrefix(key, "where "), value)
			return true
		}
		if strings.HasPrefix(key, "or ") {
			query = query.Or(strings.TrimPrefix(key, "or "), value)
			return true
		}
		if strings.HasPrefix(key, "not ") {
			query = query.Not(strings.TrimPrefix(key, "not "), value)
			return true
		}
		if strings.HasPrefix(key, "limit ") {
			query = query.Limit(value.(int))
			return true
		}
		if strings.HasPrefix(key, "offset ") {
			query = query.Offset(value.(int))
			return true
		}
		if strings.HasPrefix(key, "asc ") {
			query = query.Order(fmt.Sprintf("%s ASC", strings.TrimPrefix(key, "asc ")))
			return true
		}
		if strings.HasPrefix(key, "desc ") {
			query = query.Order(fmt.Sprintf("%s DESC", strings.TrimPrefix(key, "desc ")))
			return true
		}
		// fmt.Println(key, value)
		return true
	})
	return query
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
			field := t.Field(i)
			fieldType := t.Type().Field(i)

			// 如果是嵌套结构体(匿名字段)，递归处理
			if field.Kind() == reflect.Struct && fieldType.Anonymous {
				if !field.IsZero() {
					nestedValue := field.Interface()
					if nestedID := m.getID(nestedValue); nestedID != "" {
						return nestedID
					}
				} else {
					// 即使嵌套结构体为空，也要检查其类型定义
					nestedStruct := reflect.New(field.Type()).Elem().Interface()
					if nestedID := m.getID(nestedStruct); nestedID != "" {
						// 但由于值为空，实际上没有可用的ID
						continue
					}
				}
				continue
			}

			if field.IsZero() {
				continue
			}
			v, ok := fieldType.Tag.Lookup("gorm")
			if ok {
				for _, v2 := range strings.Split(v, ";") {
					if strings.HasPrefix(v2, "id") || strings.HasPrefix(v2, "column:id") || strings.HasPrefix(v2, "primaryKey") {
						return fmt.Sprint(field.Interface())
					}
				}
			}
		}
	}
	return
}

func (m *Model) saveOplist(mode types.WhereMode, column string, value any) {
	switch mode {
	case types.WhereIs:
		m.upsertOp.Store(column, value)
		m.OpList.Store(fmt.Sprintf("where `%s` = ?", column), value)
	case types.WhereNot:
		m.OpList.Store(fmt.Sprintf("not `%s` = ?", column), value)
	case types.WhereGt:
		m.OpList.Store(fmt.Sprintf("where `%s` > ?", column), value)
	case types.WhereLt:
		m.OpList.Store(fmt.Sprintf("where `%s` < ?", column), value)
	case types.WhereOr:
		m.OpList.Store(fmt.Sprintf("or `%s` = ?", column), value)
	case types.OrderAsc:
		m.OpList.Store(fmt.Sprintf("asc `%s`", column), "")
	case types.OrderDesc:
		m.OpList.Store(fmt.Sprintf("desc `%s`", column), "")
	case types.WhereGte:
		m.OpList.Store(fmt.Sprintf("where `%s` >= ?", column), value)
	case types.WhereLte:
		m.OpList.Store(fmt.Sprintf("where `%s` <= ?", column), value)
	case types.WhereLike:
		m.OpList.Store(fmt.Sprintf("where `%s` like ?", column), "%"+fmt.Sprint(value)+"%")
	}
}

func (m *Model) ResetFilter() types.ORMModel {
	m.OpList = types.NewOrderedMap()
	m.upsertOp = sync.Map{}
	m.Data = nil
	return m
}
