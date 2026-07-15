package sqlorm

import (
	"fmt"

	"gorm.io/gorm"
)

// Incr 将当前 Where 条件匹配的记录中 column 原子增加 amount。
// 生成 SQL: UPDATE table SET column = column + amount WHERE ...
// 使用 gorm.Expr 保证表达式原样写入，不被参数化成值。
func (m *Model) Incr(column string, amount int64) error {
	return m.makeQuery().
		UpdateColumn(column, gorm.Expr(fmt.Sprintf("%s + ?", column), amount)).Error
}

// UpdateColumns 用 map 原样更新列，不跳过零值。
// data 可以是 map[string]any 或结构体。
// 当 data 里包含 gorm.Expr 时会原样展开为 SQL 表达式（如 view_count + 1）。
func (m *Model) UpdateColumns(data any) error {
	return m.makeQuery().UpdateColumns(data).Error
}

