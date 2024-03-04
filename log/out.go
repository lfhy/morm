package log

import (
	"context"
	"errors"
	"fmt"
)

var ctx = context.Background()

// 错误
func Errorln(v ...any) error {
	GetDBLoger().Error(ctx, "", fmt.Sprint(v...))
	return errors.New(fmt.Sprint(v...))
}

func Error(v ...any) error {
	return Errorln(v...)
}

func Errorf(format string, v ...any) error {
	GetDBLoger().Error(ctx, format, v...)
	return fmt.Errorf(format, v...)
}

// 调试
func Debugf(format string, v ...any) {
	GetDBLoger().Info(ctx, format, v...)
}
