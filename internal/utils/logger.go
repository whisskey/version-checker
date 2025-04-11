package utils

import (
	"fmt"

	"version-checker/config"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Options struct {
	Level             string
	Development       bool
	DisableCaller     bool
	DisableStacktrace bool
}

func New(cfg *config.Config) (*zap.Logger, error) {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    "func",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var level zapcore.Level
	if err := level.UnmarshalText([]byte(cfg.Logger.Level)); err != nil {
		return nil, fmt.Errorf("invalid log level %q: %v", cfg.Logger.Level, err)
	}

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(level),
		Development:       false,
		Encoding:          "console",
		EncoderConfig:     encoderConfig,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
		DisableCaller:     true,
		DisableStacktrace: true,
	}

	logger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %v", err)
	}

	return logger, nil
}

func NewLogger(cfg *config.Config) *zap.Logger {
	logger, err := New(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}
	return logger
}

var Module = fx.Options(
	fx.Provide(NewLogger),
)
