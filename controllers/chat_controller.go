package controllers

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"pet-wellness-backend/services"
)

// ChatController handles HTTP requests for the /chat endpoints.
type ChatController struct {
	service *services.ChatService
}

func NewChatController(service *services.ChatService) *ChatController {
	return &ChatController{service: service}
}

func (c *ChatController) SendMessage(ctx *fiber.Ctx) error {
	var req struct {
		UserID  string `json:"user_id"`
		Message string `json:"message"`
	}
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "invalid request body",
		})
	}
	if req.UserID == "" || req.Message == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "user_id and message are required",
		})
	}

	reply, err := c.service.SendMessage(context.Background(), req.UserID, req.Message)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "error",
				"message": "pet not found, please set up your pet first",
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "failed to process chat message",
		})
	}

	return ctx.JSON(fiber.Map{
		"status": "success",
		"reply":  reply,
	})
}

func (c *ChatController) GetHistory(ctx *fiber.Ctx) error {
	userID := ctx.Params("user_id")
	if userID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "user_id is required",
		})
	}

	messages, err := c.service.GetHistory(userID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "failed to fetch chat history",
		})
	}

	return ctx.JSON(fiber.Map{
		"status": "success",
		"data":   messages,
	})
}
