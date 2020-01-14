package logger

import (
	"log"

	"go.uber.org/zap"
)

// NewLogger : creates a new logger instance
func NewLogger() *zap.Logger {
	logger, err := zap.NewProduction()

	if err != nil {
		log.Panic(err)
	}

	return logger
}
