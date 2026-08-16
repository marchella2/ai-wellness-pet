package routes

import (
	"github.com/gofiber/fiber/v2"

	"pet-wellness-backend/controllers"
)

// UserRouter registers the /user endpoints on the Fiber app.
type UserRouter struct {
	app        *fiber.App
	controller *controllers.UserController
}

func NewUserRouter(app *fiber.App, controller *controllers.UserController) *UserRouter {
	return &UserRouter{app: app, controller: controller}
}

func (r *UserRouter) Register() {
	api := r.app.Group("/api/v1")
	api.Post("/user/register", r.controller.Register)
}
