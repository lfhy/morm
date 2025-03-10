package mongodb

import (
	"fmt"
	"reflect"
	"strings"

	orm "github.com/lfhy/morm/interface"
	"github.com/lfhy/morm/log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	m.WhereList[key] = value
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
	m.OpList.Store("limit", int64(limit))
	return m
}

// 跳过查询的数量
func (m Model) Offset(offset int) orm.ORMModel {
	m.OpList.Store("offset", int64(offset))
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
	if t.Kind() == reflect.Ptr {
		if t.IsNil() {
			t = reflect.New(t.Type())
		}
		t = t.Elem()
	}
	if t.Kind() == reflect.Slice {
		t = t.Elem()
	}

	if m.WhereList == nil {
		m.WhereList = bson.M{}
	}

	switch t.Kind() {
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			if t.Field(i).IsZero() {
				continue
			}
			dtype := t.Type()
			value := dtype.Field(i)
			v, ok := value.Tag.Lookup("bson")
			if ok {
				switch mode {
				case WhereIs:
					if v == "_id" {
						m.WhereList[v] = t.Field(i).Interface()
						continue
					}
					m.WhereList[v] = bson.M{"$eq": t.Field(i).Interface()}
				case WhereNot:
					if v == "_id" {
						ids, err := primitive.ObjectIDFromHex(t.Field(i).Interface().(string))
						if err != nil {
							log.Error(err)
						}
						m.WhereList[v] = bson.M{"$ne": ids}
						continue
					}
					m.WhereList[v] = bson.M{"$ne": t.Field(i).Interface()}
				case WhereGt:
					m.WhereList[v] = bson.M{"$gt": t.Field(i).Interface()}
				case WhereLt:
					m.WhereList[v] = bson.M{"$lt": t.Field(i).Interface()}
				case WhereOr:
					m.WhereList[v] = bson.M{"$or": t.Field(i).Interface()}
				case OrderAsc:
					data, ok := m.OpList.Load("asc")
					if !ok {
						m.OpList.Store("asc", bson.D{{Key: v, Value: 1}})
					} else {
						sort := data.(bson.D)
						sort = append(sort, bson.E{Key: v, Value: 1})
						m.OpList.Store("asc", sort)
					}
				case OrderDesc:
					data, ok := m.OpList.Load("desc")
					if !ok {
						m.OpList.Store("desc", bson.D{{Key: v, Value: -1}})
					} else {
						sort := data.(bson.D)
						sort = append(sort, bson.E{Key: v, Value: -1})
						m.OpList.Store("desc", sort)
					}
				}

			}
		}
	}
	if m.WhereList["_id"] != nil {
		if _, ok := m.WhereList["_id"].(primitive.ObjectID); !ok {
			// 如果不是 primitive.M 类型，进行转换
			ids, err := primitive.ObjectIDFromHex(fmt.Sprint(m.WhereList["_id"]))
			if err != nil {
				log.Error(err)
			}
			m.WhereList["_id"] = ids
		}

	}
	return m
}

func (m Model) makeQuary() options.FindOptions {

	opts := options.Find()
	if m.OpList != nil {
		m.OpList.Range(func(key, value interface{}) bool {
			if strings.HasPrefix(key.(string), "limit") {
				opts = opts.SetLimit(value.(int64))
				return true
			}
			if strings.HasPrefix(key.(string), "offset") {
				opts = opts.SetSkip(value.(int64))
				return true
			}
			if strings.HasPrefix(key.(string), "asc") {
				opts = opts.SetSort(value)
				return true
			}
			if strings.HasPrefix(key.(string), "desc") {
				opts = opts.SetSort(value)
				return true
			}
			return true
		})

	}
	return *opts
}

func setIDField(dataStruct interface{}, value string) {
	val := reflect.ValueOf(dataStruct)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			val = reflect.New(val.Type())
		}
		val = val.Elem()
	}
	idField := val.FieldByName("ID")
	if idField.IsValid() && idField.CanSet() {
		idField.Set(reflect.ValueOf(value))
	}
}
