package routes

import (
	"github.com/gofiber/fiber/v2"

	"pet-wellness-backend/controllers"
)

// PetRouter registers the /pet endpoints on the Fiber app.
type PetRouter struct {
	app        *fiber.App
	controller *controllers.PetController
}

func NewPetRouter(app *fiber.App, controller *controllers.PetController) *PetRouter {
	return &PetRouter{app: app, controller: controller}
}

func (r *PetRouter) Register() {
	api := r.app.Group("/api/v1")
	api.Get("/pet/:user_id", r.controller.GetByUserID)
	api.Post("/pet/setup", r.controller.Setup)
	api.Post("/pet/:user_id/reset", r.controller.Reset)
	api.Post("/pet/:user_id/simulate-neglect", r.controller.SimulateNeglect)
}
