package types

type DBType string

const (
	MySQL   DBType = "mysql"
	MongoDB DBType = "mongodb"
	SQLite  DBType = "sqlite"
)
