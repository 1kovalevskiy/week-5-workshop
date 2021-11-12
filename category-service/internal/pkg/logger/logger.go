package logger

import (
	"context"
	"log"

	"go.uber.org/zap"
)

type ctxKey struct{}

var attachedLoggerKey = &ctxKey{}

var globalLogger *zap.SugaredLogger

func fromContext(ctx context.Context) *zap.SugaredLogger {
	if attachedLogger, ok := ctx.Value(attachedLoggerKey).(*zap.SugaredLogger); ok {
		return attachedLogger
	}

	return globalLogger
}

func ErrorKV(ctx context.Context, message string, kvs ...interface{}) {
	fromContext(ctx).Error(append([]interface{}{message}, kvs...)...)
}

func WarnKV(ctx context.Context, message string, kvs ...interface{}) {
	fromContext(ctx).Warn(append([]interface{}{message}, kvs...)...)
}

func InfoKV(ctx context.Context, message string, kvs ...interface{}) {
	fromContext(ctx).Info(append([]interface{}{message}, kvs...)...)
}

func DebugKV(ctx context.Context, message string, kvs ...interface{}) {
	fromContext(ctx).Debug(append([]interface{}{message}, kvs...)...)
}

func FatalKV(ctx context.Context, message string, kvs ...interface{}) {
	fromContext(ctx).Fatal(append([]interface{}{message}, kvs...)...)
}

func AttachLogger(ctx context.Context, logger *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, attachedLoggerKey, logger)
}

func SetLogger(newLogger *zap.SugaredLogger) {
	globalLogger = newLogger
}

func init() {
	notSugaredLogger, err := zap.NewProduction()
	if err != nil {
		log.Panic(err)
	}

	globalLogger = notSugaredLogger.Sugar()
}
