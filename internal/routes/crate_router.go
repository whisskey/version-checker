package routes

import (
	"version-checker/internal/handler"

	"github.com/gofiber/fiber/v2"
)

type Router struct {
	handler handler.CrateHandler
}

func NewCrate(h handler.CrateHandler) *Router {
	return &Router{
		handler: h,
	}
}

func (r *Router) Setup(app fiber.Router) {
	crates := app.Group("/crates")
	crates.Get("/:name/:version", r.handler.CheckVersion)
}
