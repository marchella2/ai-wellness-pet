package controllers

import (
	"github.com/gofiber/fiber/v2"

	"pet-wellness-backend/services"
)

// HealthController handles HTTP requests for the /health endpoint.
type HealthController struct {
	service *services.HealthService
}

func NewHealthController(service *services.HealthService) *HealthController {
	return &HealthController{service: service}
}

func (c *HealthController) Check(ctx *fiber.Ctx) error {
	return ctx.JSON(c.service.Check())
}
