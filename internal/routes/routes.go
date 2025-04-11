package routes

import (
	"version-checker/internal/handler"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
)

type Route struct {
	Fiber *fiber.App
}

func NewRoute(app *fiber.App) Route {
	return Route{
		Fiber: app,
	}
}

var Module = fx.Options(
	fx.Provide(NewRoute),
	fx.Invoke(registerRoutes),
)

func registerRoutes(
	crateController handler.CrateHandler,
	route Route,
) {
	api := route.Fiber.Group("/api")

	crateRouter := NewCrate(crateController)

	crateRouter.Setup(api)
}
