package controllers

import (
	"errors"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"pet-wellness-backend/services"
)

// PetController handles HTTP requests for the /pet endpoints.
type PetController struct {
	service *services.PetService
}

func NewPetController(service *services.PetService) *PetController {
	return &PetController{service: service}
}

func (c *PetController) GetByUserID(ctx *fiber.Ctx) error {
	userID := ctx.Params("user_id")
	if userID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "user_id is required",
		})
	}

	pet, err := c.service.GetByUserID(userID)
	if err != nil {
		return c.respondPetError(ctx, err, "failed to fetch pet")
	}

	return ctx.JSON(fiber.Map{
		"status": "success",
		"pet":    pet,
	})
}

func (c *PetController) Setup(ctx *fiber.Ctx) error {
	var req struct {
		UserID  string `json:"user_id"`
		PetName string `json:"pet_name"`
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

	petName := strings.TrimSpace(req.PetName)
	if petName == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "pet_name is required",
		})
	}

	pet, err := c.service.Setup(req.UserID, petName)
	if err != nil {
		if errors.Is(err, services.ErrUserNotRegistered) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "error",
				"message": "user not registered, please register first",
			})
		}
		log.Printf("[pet] failed to setup pet: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "failed to setup pet",
		})
	}

	return ctx.JSON(fiber.Map{
		"status":  "success",
		"message": "Pet initialized successfully",
		"pet":     pet,
	})
}

func (c *PetController) Reset(ctx *fiber.Ctx) error {
	userID := ctx.Params("user_id")
	if userID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "user_id is required",
		})
	}

	pet, err := c.service.Reset(userID)
	if err != nil {
		return c.respondPetError(ctx, err, "failed to reset pet")
	}

	return ctx.JSON(fiber.Map{
		"status":  "success",
		"message": "Pet state reset to default",
		"pet":     pet,
	})
}

func (c *PetController) SimulateNeglect(ctx *fiber.Ctx) error {
	userID := ctx.Params("user_id")
	if userID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "user_id is required",
		})
	}

	pet, err := c.service.SimulateNeglect(userID)
	if err != nil {
		return c.respondPetError(ctx, err, "failed to simulate neglect")
	}

	return ctx.JSON(fiber.Map{
		"status":  "success",
		"message": "Pet is now neglected/sad",
		"pet":     pet,
	})
}

func (c *PetController) respondPetError(ctx *fiber.Ctx, err error, failureMessage string) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "pet not found",
		})
	}
	return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"status":  "error",
		"message": failureMessage,
	})
}
