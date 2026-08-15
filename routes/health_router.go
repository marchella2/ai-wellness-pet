package routes

import (
	"github.com/gofiber/fiber/v2"

	"pet-wellness-backend/controllers"
)

// HealthRouter registers the /health endpoint on the Fiber app.
type HealthRouter struct {
	app        *fiber.App
	controller *controllers.HealthController
}

func NewHealthRouter(app *fiber.App, controller *controllers.HealthController) *HealthRouter {
	return &HealthRouter{app: app, controller: controller}
}

func (r *HealthRouter) Register() {
	r.app.Get("/health", r.controller.Check)
}
