package handler

import (
	"version-checker/internal/service"
	"version-checker/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type CrateHandler interface {
	CheckVersion(c *fiber.Ctx) error
}

type Crate struct {
	svc service.CrateService
}

func New(svc service.CrateService) CrateHandler {
	return &Crate{
		svc: svc,
	}
}

func (h *Crate) CheckVersion(c *fiber.Ctx) error {
	name := c.Params("name")
	version := c.Params("version")

	result, err := h.svc.CheckVersion(name, version)
	if err != nil {
		if err.Error() == "package not found" {
			return utils.NewNotFoundError(err.Error())
		}
		return utils.NewBadRequestError(err.Error())
	}

	return utils.SendResponse(c, fiber.StatusOK, utils.SuccessResponse(fiber.Map{
		"package_name":    name,
		"current_version": version,
		"latest_version":  result.Latest,
		"has_update":      result.HasUpdate,
	}))
}
