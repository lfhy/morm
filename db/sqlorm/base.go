package sqlorm

import (
	"errors"
	"fmt"

	"github.com/lfhy/morm/types"
)

// 插入数据
func (m *Model) Create(data any) (err error) {
	if data != nil {
		m.Data = data
	}
	err = m.getDB().Create(m.Data).Scan(m.Data).Error
	return
}

func (m *Model) Insert(data any) (err error) {
	return m.Create(data)
}

// 更新或插入数据
func (m *Model) Save(data any, value ...any) (err error) {
	q := m.makeQuary()
	if len(value) > 0 {
		if col, ok := data.(string); ok {
			q.Set(col, value[0])
		}
	} else {
		if data != nil {
			m.Data = data
		}
	}
	var i int64
	q.Count(&i)
	if i > 0 {
		return m.Update(data, value...)
	} else {
		// 组合新数据
		m.whereMode(data, WhereIs)
		newData := make(map[string]any)
		m.upsertOp.Range(func(key, value any) bool {
			newData[key.(string)] = value
			return true
		})
		table := m.Table
		if table == "" {
			table = GetTableName(data)
		}
		tx := m.getDB().Table(table).Create(newData)
		if _, ok := data.(string); ok {
			return tx.Error
		}
		return tx.Scan(data).Error
	}
}

func (m *Model) Upsert(data any, value ...any) error {
	return m.Save(data, value...)
}

// 删除
func (m *Model) Delete(data ...any) error {
	if len(data) > 0 && data[0] != nil {
		m.Data = data[0]
	}
	return m.makeQuary().Delete(m.Data).Error
}

// 修改
func (m *Model) Update(data any, value ...any) error {
	if len(value) > 0 {
		col, ok := data.(string)
		if ok {
			return m.makeQuary().Update(col, value[0]).Error
		}
	}
	if data != nil {
		m.Data = data
	}
	return m.makeQuary().Updates(m.Data).Error
}

// 查询数据
func (m *Model) Find() types.ORMQuary {
	return &Quary{m: m, OpList: &m.OpList}
}

func (q *Model) One(data any) error {
	return q.Find().One(data)
}

func (q *Model) All(data any) error {
	return q.Find().All(data)
}

func (q *Model) Count() int64 {
	return q.Find().Count()
}

func (q *Model) Cursor() (types.Cursor, error) {
	return q.Find().Cursor()

}

/*
**

	operations := []BulkWriteOperation{
	    {
	        Type: "insert",
	        Data: &User{Name: "Alice"},
	    },
	    {
	        Type:  "update",
	        Data:  &User{},
	        Where: map[string]any{"name": "Bob"},
	        Values: map[string]any{"age": 30},
	    },
	    {
	        Type:  "delete",
	        Data:  &User{},
	        Where: map[string]any{"name": "Charlie"},
	    },
	}

err := model.BulkWrite(operations, true)
**
*/
func (m *Model) BulkWrite(datas any, order bool) error {
	operations, ok := datas.([]types.BulkWriteOperation)
	if !ok {
		return errors.New("datas must be []orm.BulkWriteOperation")
	}

	if len(operations) == 0 {
		return nil
	}

	tx := m.getDB().Begin()
	if tx.Error != nil {
		return tx.Error
	}

	for _, op := range operations {
		switch op.Type {
		case "insert":
			if err := tx.Create(op.Data).Error; err != nil {
				if order {
					tx.Rollback()
					return err
				}
				continue
			}
		case "update":
			q := tx.Model(op.Data)
			if len(op.Where) > 0 {
				q = q.Where(op.Where)
			}
			if err := q.Updates(op.Values).Error; err != nil {
				if order {
					tx.Rollback()
					return err
				}
				continue
			}
		case "delete":
			q := tx.Model(op.Data)
			if len(op.Where) > 0 {
				q = q.Where(op.Where)
			}
			if err := q.Delete(op.Data).Error; err != nil {
				if order {
					tx.Rollback()
					return err
				}
				continue
			}
		default:
			if order {
				tx.Rollback()
				return fmt.Errorf("unsupported operation type: %s", op.Type)
			}
			continue
		}
	}

	return tx.Commit().Error
}
