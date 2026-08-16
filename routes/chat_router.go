package routes

import (
	"github.com/gofiber/fiber/v2"

	"pet-wellness-backend/controllers"
)

// ChatRouter registers the /chat endpoints on the Fiber app.
type ChatRouter struct {
	app        *fiber.App
	controller *controllers.ChatController
}

func NewChatRouter(app *fiber.App, controller *controllers.ChatController) *ChatRouter {
	return &ChatRouter{app: app, controller: controller}
}

func (r *ChatRouter) Register() {
	api := r.app.Group("/api/v1")
	api.Post("/chat", r.controller.SendMessage)
	api.Get("/chat/:user_id", r.controller.GetHistory)
}
