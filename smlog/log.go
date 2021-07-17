package smlog

import (
	"context"
	zap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)
//Key is of type int
type key int

const logKey key =1

var (
	ZapLoggerConfig *zap.Config
)

//Init to initialize the logger
func Init(aInctx context.Context, serviceName string) *zap.Logger{
	l:=NewLogger(aInctx,serviceName)
	return l
}

//NewLogger to return the logger
func NewLogger(aInctx context.Context, appName string) *zap.Logger {
	if ZapLoggerConfig == nil {
		ZapLoggerConfig = &zap.Config{
			Level:       zap.NewAtomicLevel(),
			Sampling: &zap.SamplingConfig{
				Initial:    100,
				Thereafter: 100,
			},
			Encoding: "json",
			EncoderConfig: zapcore.EncoderConfig{
				TimeKey:        "ts",
				LevelKey:       "level",
				NameKey:        "logger",
				CallerKey:      "caller",
				MessageKey:     "msg",
				StacktraceKey:  "stacktrace",
				EncodeLevel:    zapcore.LowercaseLevelEncoder,
				EncodeTime:     zapcore.ISO8601TimeEncoder,
				EncodeDuration: zapcore.SecondsDurationEncoder,
			},
			OutputPaths:      []string{"stderr"},
			ErrorOutputPaths: []string{"stderr"},
		}

	}
	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(ZapLoggerConfig.EncoderConfig),
		nil,//TODO: Pass file loc
		ZapLoggerConfig.Level,
	))
  return logger
}

// ContextWithValue returns a new Context that carries the specified Logger.
func ContextWithValue(aInCtx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(aInCtx, logKey, logger)
}

func MustFromContext(aInCtx context.Context) *zap.Logger{
	if aInCtx == nil {
		panic("Log Initialization Failed")
	}
	logger, ok := aInCtx.Value(logKey).(*zap.Logger)
	if !ok{
		panic("Log Initialization Failed")
	}
	return logger
}