package sqlorm

import orm "github.com/lfhy/morm/interface"

// 插入数据
func (m Model) Create(data interface{}) (id string, err error) {
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
func (m Model) Save(data interface{}) (id string, err error) {
	if data != nil {
		m.Data = data
	}
	err = m.makeQuary().Save(m.Data).Scan(m.Data).Error
	if err == nil {
		id = m.getID(m.Data)
	}
	return
}

// 删除
func (m Model) Delete(data interface{}) error {
	if data != nil {
		m.Data = data
	}
	return m.makeQuary().Delete(m.Data).Error
}

// 修改
func (m Model) Update(data interface{}) error {
	if data != nil {
		m.Data = data
	}
	return m.makeQuary().Updates(m.Data).Error
}

// 查询数据
func (m Model) Find() orm.ORMQuary {
	return Quary{m: m, OpList: m.OpList}
}
