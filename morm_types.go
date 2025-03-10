package morm

import (
	orm "github.com/lfhy/morm/interface"
)

// 为了避免识别错误设置Bool类型为int类型
type BoolORM = orm.BoolORM

const (
	TrueInt  BoolORM = 1
	FalseInt BoolORM = -1
	// 空字符串
	EmptyStr = "-"
)

type ORM = orm.ORM

type ORMModel = orm.ORMModel

type ORMQuary = orm.ORMQuary
