package sqlorm

import (
	"errors"
	"fmt"

	orm "github.com/lfhy/morm/interface"
)

// 插入数据
func (m Model) Create(data any) (id string, err error) {
	if data != nil {
		m.Data = data
	}
	err = m.tx.getDB().Create(m.Data).Scan(m.Data).Error
	if err == nil {
		id = m.getID(m.Data)
	}
	return
}

// 更新或插入数据
func (m Model) Save(data any, value ...any) (id string, err error) {
	if data != nil {
		m.Data = data
	}
	q := m.makeQuary()
	if len(value) > 0 {
		if col, ok := data.(string); ok {
			q.Set(col, value[0])
		}
	}
	err = q.Save(m.Data).Scan(m.Data).Error
	if err == nil {
		id = m.getID(m.Data)
	}
	return
}

// 删除
func (m Model) Delete(data any) error {
	if data != nil {
		m.Data = data
	}
	return m.makeQuary().Delete(m.Data).Error
}

// 修改
func (m Model) Update(data any, value ...any) error {
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
func (m Model) Find() orm.ORMQuary {
	return Quary{m: m, OpList: m.OpList}
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
func (m Model) BulkWrite(datas any, order bool) error {
	operations, ok := datas.([]orm.BulkWriteOperation)
	if !ok {
		return errors.New("datas must be []orm.BulkWriteOperation")
	}

	if len(operations) == 0 {
		return nil
	}

	tx := m.tx.getDB().Begin()
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
