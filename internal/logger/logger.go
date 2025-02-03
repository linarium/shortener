package logger

import (
	"go.uber.org/zap"
	"log"
)

var Sugar *zap.SugaredLogger = zap.NewNop().Sugar()

func Initialize(opts ...zap.Option) {
	cfg := zap.NewDevelopmentConfig()

	logger, err := cfg.Build(opts...)
	if err != nil {
		log.Fatalf("Ошибка при инициализации логгера: %v", err)
	}

	Sugar = logger.Sugar()
}
