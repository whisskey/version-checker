package bootstrap

import (
	"version-checker/cmd/server"
	"version-checker/config"
	"version-checker/internal/utils"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

var Module = fx.Options(
	config.Module,
	utils.Module,
	fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
		return &fxevent.ZapLogger{Logger: log.Named("fx").WithOptions(zap.IncreaseLevel(zap.ErrorLevel))}
	}),
	server.Module,
)
