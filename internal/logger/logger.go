package logger

import (
	"go.uber.org/zap"
	"log"
)

var Sugar *zap.SugaredLogger

func Initialize() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Ошибка при инициализации логгера: %v", err)
	}

	Sugar = logger.Sugar()
}

func Sync() {
	if Sugar != nil {
		_ = Sugar.Sync()
	}
}
