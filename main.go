package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"pet-wellness-backend/config"
	"pet-wellness-backend/controllers"
	"pet-wellness-backend/routes"
	"pet-wellness-backend/services"
)

func main() {
	env := config.LoadEnv()
	db := config.ConnectDatabase(env.DatabaseURL)

	app := fiber.New(fiber.Config{
		AppName:      "AI Wellness Pet API",
		ErrorHandler: fiber.DefaultErrorHandler,
	})

	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	userService := services.NewUserService(db) // NEW
	petService := services.NewPetService(db)
	activityService := services.NewActivityService(db, env, petService)

	routes.NewHealthRouter(app, controllers.NewHealthController(services.NewHealthService())).Register()
	routes.NewUserRouter(app, controllers.NewUserController(userService)).Register() // NEW
	routes.NewPetRouter(app, controllers.NewPetController(petService)).Register()
	routes.NewActivityRouter(app, controllers.NewActivityController(activityService)).Register()

	port := env.Port
	log.Printf("[main] AI Wellness Pet API is running on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("[main] failed to start server: %v", err)
	}
}
