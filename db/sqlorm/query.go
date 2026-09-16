package sqlorm

import (
	"database/sql"

	"github.com/lfhy/morm/log"

	"github.com/lfhy/morm/types"
	"gorm.io/gorm"
)

type Query struct {
	m      *Model
	OpList *types.OrderedMap
}

func (q *Query) One(data any) error {
	return q.m.makeQuery().First(data).Error
}

func (q *Query) All(data any) error {
	return q.m.makeQuery().Find(data).Error
}

func (q *Query) Count() int64 {
	i, err := q.CountWithError()
	if err != nil {
		log.Errorf("Mysql查出错: %v\n", err)
	}
	return i
}

// CountWithError 返回 SQL 查询个数及执行错误。
func (q *Query) CountWithError() (int64, error) {
	var i int64
	err := q.m.makeQuery().Count(&i).Error
	return i, err
}

func (q *Query) Delete() error {
	return q.m.makeQuery().Delete(q.m.Data).Error
}

// gorm不支持游标，使用原始SQL实现
func (q *Query) Cursor() (types.Cursor, error) {
	rows, err := q.m.makeQuery().Rows()
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

// Err 返回 SQL 游标迭代期间产生的错误。
func (c *Cursor) Err() error {
	return c.Rows.Err()
}

func (c *Cursor) Decode(v any) error {
	err := c.db.ScanRows(c.Rows, v)
	if err != nil {
		log.Errorf("Mysql游标解码出错: %v\n", err)
	}
	return err
}
