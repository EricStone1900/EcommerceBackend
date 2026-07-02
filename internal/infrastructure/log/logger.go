package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// InitLogger creates a zap.Logger based on the environment setting.
// In development mode, it uses a human-readable console format.
// In production mode, it uses structured JSON output.
func InitLogger(env string) (*zap.Logger, error) {
	var cfg zap.Config

	if env == "production" {
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "timestamp"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
		cfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	}

	cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	if env != "production" {
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	zap.ReplaceGlobals(logger)
	return logger, nil
}
