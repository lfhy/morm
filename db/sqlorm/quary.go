package sqlorm

import (
	"database/sql"

	"github.com/lfhy/morm/log"

	"github.com/lfhy/morm/types"
	"gorm.io/gorm"
)

type Quary struct {
	m      *Model
	OpList *types.OrderedMap
}

func (q *Quary) One(data any) error {
	return q.m.makeQuary().First(data).Error
}

func (q *Quary) All(data any) error {
	return q.m.makeQuary().Find(data).Error
}

func (q *Quary) Count() int64 {
	var i int64
	q.m.makeQuary().Count(&i)
	return i
}

func (q *Quary) Delete() error {
	return q.m.makeQuary().Delete(q.m.Data).Error
}

// gorm不支持游标，使用原始SQL实现
func (q *Quary) Cursor() (types.Cursor, error) {
	rows, err := q.m.makeQuary().Rows()
	if err != nil {
		log.Errorf("Mysql查出错: %v\n", err)
		return nil, err
	}
	return &Cursor{Rows: rows, db: q.m.getDB()}, nil
}

type Cursor struct {
	db *gorm.DB
	*sql.Rows
}

func (c *Cursor) Decode(v any) error {
	err := c.db.ScanRows(c.Rows, v)
	if err != nil {
		log.Errorf("Mysql游标解码出错: %v\n", err)
	}
	return err
}
