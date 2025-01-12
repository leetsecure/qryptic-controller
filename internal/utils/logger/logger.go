package logger

import (
	"context"
	"os"

	"go.uber.org/zap"
)

type LoggerKeyType int

const loggerKey LoggerKeyType = iota

var logger *zap.SugaredLogger

func NewContext(
	ctx context.Context,
	serviceName string,
) context.Context {
	return context.WithValue(
		ctx,
		loggerKey,
		WithContext(ctx).With(
			"serviceName", serviceName),
	)
}

func WithContext(ctx context.Context) *zap.SugaredLogger {
	if ctx == nil {
		return logger
	}

	if ctxLogger, ok := ctx.Value(loggerKey).(*zap.SugaredLogger); ok {
		return ctxLogger
	}

	return logger
}

func Default() *zap.SugaredLogger {
	return logger
}

func setLogger(l *zap.SugaredLogger) {
	logger = l
}

func LogBuildVersionNumber() {
	if logger == nil {
		return
	}

	buildVersion := os.Getenv("BUILD_VERSION")
	if buildVersion == "" {
		return
	}

	// Log the build version
	logger.Infoln("Build version:", buildVersion)
}

func init() {
	env, ok := os.LookupEnv("LOG_ENV")
	if !ok {
		env = "development"
	}

	var cfg zap.Config

	switch env {
	case "development":
		cfg = zap.NewDevelopmentConfig()
	case "production":
		cfg = zap.NewProductionConfig()
	default:
		cfg = zap.NewDevelopmentConfig()
	}

	baseLogger, _ := cfg.Build()

	logger = baseLogger.Sugar()

	LogBuildVersionNumber()
}
