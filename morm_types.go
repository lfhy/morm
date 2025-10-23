package morm

import (
	"github.com/lfhy/morm/conf"
	"github.com/lfhy/morm/types"
)

// 为了避免识别错误设置Bool类型为int类型
type BoolORM = types.BoolORM

const (
	BoolORMTrue  BoolORM = 1
	BoolORMFalse BoolORM = -1
	// 空字符串
	EmptyStr = "-"
)

type ORM = types.ORM

type ORMModel = types.ORMModel

type ORMQuary = types.ORMQuary

type BulkWriteOperation = types.BulkWriteOperation

type MongoBulkWriteOperation = types.MongoBulkWriteOperation

type DBConfig = conf.DBConfig

type SQLiteConfig = conf.SQLiteConfig

type MySQLConfig = conf.MySQLConfig

type MongoDBConfig = conf.MongoDBConfig

type Session = types.Session
