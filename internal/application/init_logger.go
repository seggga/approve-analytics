package application

import (
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// initLogger creates a logger with additional field "service"
func initLogger(lvl string) *zap.Logger {

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "time"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	level, err := zap.ParseAtomicLevel(lvl)
	if err != nil {
		log.Fatalf("cannot parse loglevel %s: %v", lvl, err)
	}

	logger := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.Lock(os.Stdout),
			level,
		),
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Fields(zapcore.Field{
			Key:    "service",
			Type:   zapcore.StringType,
			String: "analytics",
		}),
	)

	return logger
}
