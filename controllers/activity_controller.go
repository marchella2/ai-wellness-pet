package controllers

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"pet-wellness-backend/services"
)

// ActivityController handles HTTP requests for the /activity endpoints.
type ActivityController struct {
	service *services.ActivityService
}

func NewActivityController(service *services.ActivityService) *ActivityController {
	return &ActivityController{service: service}
}

func (c *ActivityController) Create(ctx *fiber.Ctx) error {
	var req struct {
		UserID       string  `json:"user_id"`
		WaterGlasses int     `json:"water_glasses"`
		SleepHours   float64 `json:"sleep_hours"`
		JournalText  string  `json:"journal_text"`
	}
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "invalid request body",
		})
	}
	if req.UserID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "user_id is required",
		})
	}

	pet, aiMessage, err := c.service.LogActivity(context.Background(), services.ActivityLogInput{
		UserID:       req.UserID,
		WaterGlasses: req.WaterGlasses,
		SleepHours:   req.SleepHours,
		JournalText:  req.JournalText,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "error",
				"message": "pet not found, please set up your pet first",
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "failed to process activity",
		})
	}

	return ctx.JSON(fiber.Map{
		"status":     "success",
		"pet":        pet,
		"ai_message": aiMessage,
	})
}

func (c *ActivityController) GetHistory(ctx *fiber.Ctx) error {
	userID := ctx.Params("user_id")
	if userID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "user_id is required",
		})
	}

	logs, err := c.service.GetHistory(userID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "failed to fetch activity history",
		})
	}

	return ctx.JSON(fiber.Map{
		"status": "success",
		"data":   logs,
	})
}
