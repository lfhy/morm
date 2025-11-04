package types

import "gorm.io/gorm/logger"

type LogLevel = logger.LogLevel

const (
	LogLevelSilent LogLevel = iota + 1
	LogLevelError
	LogLevelWarn
	LogLevelInfo
)
