package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type logger interface {
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
}

// Logger is used globally. This can be assigned any struct implements logger interface.
var Logger = func() logger {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	logger, _ := cfg.Build()
	return logger.Sugar()
}()
