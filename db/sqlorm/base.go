package sqlorm

import orm "github.com/lfhy/morm/interface"

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
