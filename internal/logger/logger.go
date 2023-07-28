package logger

import (
	"log"

	"go.uber.org/zap"
)

var zaplog *zap.Logger

// Конструктор нового логгера
func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	var cfg zap.Config
	if lvl.Level() == zap.DebugLevel {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}

	cfg.Level = lvl
	cfg.DisableCaller = true

	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	zaplog = zl
	return nil
}

// Отладочные сообщения
func Debug(msg string, a ...any) {
	if zaplog != nil {
		zaplog.Sugar().Debugw(msg, a...)
	} else {
		stdLogPrint("DEBUG", msg, a...)
	}
}

// Информационные сообщения
func Info(msg string, a ...any) {
	if zaplog != nil {
		zaplog.Sugar().Infow(msg, a...)
	} else {
		stdLogPrint("INFO", msg, a...)
	}
}

// Предупреждения
func Warn(msg string, a ...any) {
	if zaplog != nil {
		zaplog.Sugar().Warnw(msg, a...)
	} else {
		stdLogPrint("WARN", msg, a...)
	}
}

// Ошибки
func Error(err error, a ...any) {
	if zaplog != nil {
		zaplog.Sugar().Errorw(err.Error(), a...)
	} else {
		stdLogPrint("ERROR", err, a...)
	}
}

// вывод в стандартный лог
func stdLogPrint(level string, msg any, v ...any) {
	p := []any{level, msg}
	p = append(p, v...)
	log.Println(p...)
}
