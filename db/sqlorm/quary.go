package sqlorm

import (
	"database/sql"
	"sync"

	"github.com/lfhy/morm/types"
	"gorm.io/gorm"
)

type Quary struct {
	m      *Model
	OpList *sync.Map
}

func (q *Quary) One(data any) error {
	return q.m.makeQuary().First(data).Error
}

func (q *Quary) All(data any) error {
	return q.m.makeQuary().Find(data).Error
}

func (q *Quary) Count() (i int64, err error) {
	return i, q.m.makeQuary().Count(&i).Error
}

func (q *Quary) Delete() error {
	return q.m.makeQuary().Delete(q.m.Data).Error
}

// gorm不支持游标，使用原始SQL实现
func (q *Quary) Cursor() (types.Cursor, error) {
	rows, err := q.m.makeQuary().Rows()
	if err != nil {
		return nil, err
	}
	return &Cursor{Rows: rows, db: q.m.getDB()}, nil
}

type Cursor struct {
	db *gorm.DB
	*sql.Rows
}

func (c *Cursor) Decode(v any) error {
	return c.db.ScanRows(c.Rows, v)
}
