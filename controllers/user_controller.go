package controllers

import (
	"errors"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"

	"pet-wellness-backend/services"
)

// UserController handles HTTP requests for the /user endpoints.
type UserController struct {
	service *services.UserService
}

func NewUserController(service *services.UserService) *UserController {
	return &UserController{service: service}
}

// Register creates a new user, or -- if the email is already registered --
// returns the existing user instead of erroring. This keeps the endpoint
// idempotent for clients (e.g. the Flutter app calling this once on first
// launch, or safely retrying after a network hiccup).
func (c *UserController) Register(ctx *fiber.Ctx) error {
	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "invalid request body",
		})
	}

	name := strings.TrimSpace(req.Name)
	email := strings.TrimSpace(req.Email)
	if name == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "name is required",
		})
	}
	if email == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "email is required",
		})
	}

	user, err := c.service.Register(name, email)
	if err != nil {
		if errors.Is(err, services.ErrEmailAlreadyRegistered) {
			existing, lookupErr := c.service.GetByEmail(email)
			if lookupErr != nil {
				log.Printf("[user] email taken but lookup failed: %v", lookupErr)
				return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
					"status":  "error",
					"message": "email already registered",
				})
			}
			return ctx.JSON(fiber.Map{
				"status":  "success",
				"message": "User already registered",
				"user":    existing,
			})
		}
		log.Printf("[user] failed to register user: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "failed to register user",
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "User registered successfully",
		"user":    user,
	})
}
