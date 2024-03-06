package sqlorm

import (
	"fmt"
	"reflect"
	"strings"

	orm "github.com/lfhy/morm/interface"

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
)

// 限制条件
func (m Model) Where(condition interface{}) orm.ORMModel {
	return m.whereMode(condition, WhereIs)
}

func (m Model) WhereIs(key string, value any) orm.ORMModel {
	m.OpList.Store(fmt.Sprintf("where %s = ?", key), value)
	return m
}

func (m Model) WhereNot(condition interface{}) orm.ORMModel {
	return m.whereMode(condition, WhereNot)
}

func (m Model) WhereGt(condition interface{}) orm.ORMModel {
	return m.whereMode(condition, WhereGt)
}

func (m Model) WhereLt(condition interface{}) orm.ORMModel {
	return m.whereMode(condition, WhereLt)
}

func (m Model) WhereOr(condition interface{}) orm.ORMModel {
	return m.whereMode(condition, WhereOr)
}

// 限制查询的数量
func (m Model) Limit(limit int) orm.ORMModel {
	m.OpList.Store("limit ", limit)
	return m
}

// 跳过查询的数量
func (m Model) Offset(offset int) orm.ORMModel {
	m.OpList.Store("offset ", offset)
	return m
}

// 正序
func (m Model) Asc(condition interface{}) orm.ORMModel {
	return m.whereMode(condition, OrderAsc)
}

// 逆序
func (m Model) Desc(condition interface{}) orm.ORMModel {
	return m.whereMode(condition, OrderDesc)
}
func (m Model) whereMode(condition interface{}, mode int) orm.ORMModel {
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
							m.OpList.Store(fmt.Sprintf("where %s = ?", strings.TrimPrefix(v2, "column:")), t.Field(i).Interface())
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
func (m Model) makeQuary() *gorm.DB {
	quary := m.tx.getDB().Model(m.Data)
	if m.OpList != nil {
		
		m.OpList.Range(func(key, value interface{}) bool {
			keyStr := fmt.Sprint(key)
			if strings.HasPrefix(keyStr, "where ") {
				quary = quary.Where(strings.TrimPrefix(keyStr, "where "), value)
				return true
			}
			if strings.HasPrefix(keyStr, "or ") {
				quary = quary.Or(strings.TrimPrefix(keyStr, "or "), value)
				return true
			}
			if strings.HasPrefix(keyStr, "not ") {
				quary = quary.Not(strings.TrimPrefix(keyStr, "not "), value)
				return true
			}
			if strings.HasPrefix(keyStr, "limit ") {
				quary = quary.Limit(value.(int))
				return true
			}
			if strings.HasPrefix(keyStr, "offset ") {
				quary = quary.Offset(value.(int))
				return true
			}
			if strings.HasPrefix(keyStr, "asc ") {
				quary = quary.Order(fmt.Sprintf("%s ASC", strings.TrimPrefix(keyStr, "asc ")))
				return true
			}
			if strings.HasPrefix(keyStr, "desc ") {
				quary = quary.Order(fmt.Sprintf("%s DESC", strings.TrimPrefix(keyStr, "desc ")))
				return true
			}
			// fmt.Println(key, value)
			return true
		})
	}
	return quary
}

func (m Model) getID(condition interface{}) (id string) {
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
					if strings.HasPrefix(v2, "column:id") || strings.HasPrefix(v2, "primaryKey") {
						return fmt.Sprint(t.Field(i).Interface())
					}
				}
			}
		}
	}
	return
}
