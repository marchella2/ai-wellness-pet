package routes

import (
	"github.com/gofiber/fiber/v2"

	"pet-wellness-backend/controllers"
)

// ActivityRouter registers the /activity endpoints on the Fiber app.
type ActivityRouter struct {
	app        *fiber.App
	controller *controllers.ActivityController
}

func NewActivityRouter(app *fiber.App, controller *controllers.ActivityController) *ActivityRouter {
	return &ActivityRouter{app: app, controller: controller}
}

func (r *ActivityRouter) Register() {
	api := r.app.Group("/api/v1")
	api.Post("/activity", r.controller.Create)
	api.Get("/activity/:user_id", r.controller.GetHistory)
}
