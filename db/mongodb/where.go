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
	WhereGte
	WhereLte
)

// 限制条件
func (m Model) Where(condition any, value ...any) orm.ORMModel {
	if len(value) > 0 {
		key, ok := condition.(string)
		if ok {
			m.WhereList[key] = bson.M{"$eq": value[0]}
			return m
		}
	}
	return m.whereMode(condition, WhereIs)
}

func (m Model) WhereIs(key string, value any) orm.ORMModel {
	m.WhereList[key] = value
	return m
}

func (m Model) WhereNot(condition any, value ...any) orm.ORMModel {
	if len(value) > 0 {
		key, ok := condition.(string)
		if ok {
			m.WhereList[key] = bson.M{"$ne": value[0]}
			return m
		}
	}
	return m.whereMode(condition, WhereNot)
}

func (m Model) WhereGt(condition any, value ...any) orm.ORMModel {
	if len(value) > 0 {
		key, ok := condition.(string)
		if ok {
			m.WhereList[key] = bson.M{"$gt": value[0]}
			return m
		}
	}
	return m.whereMode(condition, WhereGt)
}

func (m Model) WhereLt(condition any, value ...any) orm.ORMModel {
	if len(value) > 0 {
		key, ok := condition.(string)
		if ok {
			m.WhereList[key] = bson.M{"$lt": value[0]}
			return m
		}
	}
	return m.whereMode(condition, WhereLt)
}

func (m Model) WhereGte(condition any, value ...any) orm.ORMModel {
	if len(value) > 0 {
		key, ok := condition.(string)
		if ok {
			m.WhereList[key] = bson.M{"$gte": value[0]}
			return m
		}
	}
	return m.whereMode(condition, WhereGte)
}

func (m Model) WhereLte(condition any, value ...any) orm.ORMModel {
	if len(value) > 0 {
		key, ok := condition.(string)
		if ok {
			m.WhereList[key] = bson.M{"$lte": value[0]}
			return m
		}
	}
	return m.whereMode(condition, WhereLte)
}

func (m Model) WhereOr(condition any, value ...any) orm.ORMModel {
	if len(value) > 0 {
		key, ok := condition.(string)
		if ok {
			m.WhereList[key] = bson.M{"$or": value[0]}
			return m
		}
	}
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
func (m Model) Asc(condition any) orm.ORMModel {
	key, ok := condition.(string)
	if ok {
		data, ok := m.OpList.Load("asc")
		if !ok {
			m.OpList.Store("asc", bson.D{{Key: key, Value: 1}})
		} else {
			sort := data.(bson.D)
			sort = append(sort, bson.E{Key: key, Value: 1})
			m.OpList.Store("asc", sort)
		}
		return m
	}
	return m.whereMode(condition, OrderAsc)
}

// 逆序
func (m Model) Desc(condition any) orm.ORMModel {
	key, ok := condition.(string)
	if ok {
		data, ok := m.OpList.Load("desc")
		if !ok {
			m.OpList.Store("desc", bson.D{{Key: key, Value: -1}})
		} else {
			sort := data.(bson.D)
			sort = append(sort, bson.E{Key: key, Value: -1})
			m.OpList.Store("desc", sort)
		}
		return m
	}
	return m.whereMode(condition, OrderDesc)
}

func (m Model) whereMode(condition any, mode int) orm.ORMModel {
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
				case WhereGte:
					m.WhereList[v] = bson.M{"$gte": t.Field(i).Interface()}
				case WhereLte:
					m.WhereList[v] = bson.M{"$lte": t.Field(i).Interface()}
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
	return m
}

func (m Model) CheckOID() {
	if m.WhereList["_id"] != nil {
		switch m.WhereList["_id"].(type) {
		case primitive.ObjectID, map[string]primitive.ObjectID:
			return
		case string:
			ids, err := primitive.ObjectIDFromHex(fmt.Sprint(m.WhereList["_id"]))
			if err != nil {
				log.Error("转换失败:", err, "原始ID:", m.WhereList["_id"])
			}
			m.WhereList["_id"] = ids
		case map[string]string:
			mp := make(map[string]primitive.ObjectID)
			for key, value := range m.WhereList["_id"].(map[string]string) {
				ids, err := primitive.ObjectIDFromHex(value)
				if err != nil {
					log.Error("转换失败:", err, "原始ID:", value)
				}
				mp[key] = ids
			}
			m.WhereList["_id"] = mp
		case bson.M:
			for key, value := range m.WhereList["_id"].(bson.M) {
				if key == "$in" {
					mp := make([]primitive.ObjectID, 0)
					for _, v := range value.([]string) {
						ids, err := primitive.ObjectIDFromHex(v)
						if err != nil {
							log.Error("转换失败:", err, "原始ID:", v)
						}
						mp = append(mp, ids)
					}
					m.WhereList["_id"].(bson.M)[key] = mp
				} else {
					ids, err := primitive.ObjectIDFromHex(fmt.Sprint(value))
					if err != nil {
						log.Error("转换失败:", err, "原始ID:", value)
					}
					m.WhereList["_id"].(bson.M)[key] = ids
				}
			}
		}
	}
}

func (m Model) makeAllQuary() options.FindOptions {
	opts := options.Find()
	if m.OpList != nil {
		m.OpList.Range(func(key, value any) bool {
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

func (m Model) makeOneQuary() options.FindOneOptions {
	opts := options.FindOne()
	if m.OpList != nil {
		m.OpList.Range(func(key, value any) bool {
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

func setIDField(dataStruct any, value string) {
	// 尝试断言为 map[string]interface{}
	if m, ok := dataStruct.(map[string]interface{}); ok {
		handleMap(m, value)
		return
	}

	// 处理结构体类型（保留原有反射逻辑）
	val := reflect.ValueOf(dataStruct)
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() == reflect.Struct {
		handleStruct(val, value)
	}
}

func handleMap(m map[string]interface{}, value string) {
	// 尝试设置 id 或 _id 键
	for _, key := range []string{"id", "_id"} {
		if _, ok := m[key]; ok {
			// 直接赋值（假设值类型兼容）
			m[key] = value
			break
		}
	}
}

// 处理 struct 类型
func handleStruct(val reflect.Value, value string) {
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		// 解析 bson 标签
		bsonTag := field.Tag.Get("bson")
		if bsonTag == "" {
			continue
		}
		// 分割标签（可能包含逗号分隔的选项）
		tagParts := strings.Split(bsonTag, ",")
		for _, part := range tagParts {
			if part == "id" || part == "_id" {
				fieldVal := val.Field(i)
				// 检查字段是否可设置且类型匹配
				if fieldVal.CanSet() && fieldVal.Kind() == reflect.String {
					fieldVal.SetString(value)
				}
				return // 找到即停止
			}
		}
	}
}
