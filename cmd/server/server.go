package server

import (
	"context"
	"fmt"

	"version-checker/config"
	"version-checker/internal/handler"
	"version-checker/internal/middleware"
	"version-checker/internal/routes"
	"version-checker/internal/service"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewFiberApp(cfg *config.Config) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler:          middleware.ErrorHandler,
	})

	return app
}

func StartFiberApp(lifecycle fx.Lifecycle, app *fiber.App, logger *zap.Logger, cfg *config.Config) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				logger.Info(fmt.Sprintf("Server starting on :%d", cfg.App.Port))

				if err := app.Listen(fmt.Sprintf(":%d", cfg.App.Port)); err != nil {
					logger.Fatal("Server failed to start", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Shutting down server")
			return app.Shutdown()
		},
	})
}

var Module = fx.Options(
	handler.Module,
	service.Module,
	routes.Module,
	fx.Provide(NewFiberApp),
	fx.Invoke(StartFiberApp),
)
