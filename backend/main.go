package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/joho/godotenv"

	"github.com/parshwanath-p2493/GreenCartt/config"
	"github.com/parshwanath-p2493/GreenCartt/handlers"
	"github.com/parshwanath-p2493/GreenCartt/utils"
)

func main() {
	// load env
	_ = godotenv.Load()
	utils.TestLoadAll()
	utils.LoadDrivers("data/drivers.xlsx")
	utils.LoadRoutes("data/routes.xlsx")
	utils.LoadOrders("data/orders.xlsx")

	port := os.Getenv("PORT")
	if port == "" {
		port = "2493"
	}

	// connect DB
	if err := config.Connect(); err != nil {
		log.Fatalf("DB connect failed: %v", err)
	}
	defer config.Disconnect()

	app := fiber.New()

	// public routes
	api := app.Group("/api")
	api.Post("/auth/register", handlers.Register)
	api.Post("/auth/login", handlers.Login)
	//api.Post("/auth/seed-manager", handlers.SeedManager) // optional seed endpoint (protected by env check)

	// CRUD public for demo (you can protect if needed)
	api.Post("/drivers", handlers.CreateDriver)
	api.Get("/drivers", handlers.GetDrivers)
	api.Get("/drivers/:id", handlers.GetDriver)
	api.Put("/drivers/:id", handlers.UpdateDriver)
	api.Delete("/drivers/:id", handlers.DeleteDriver)

	api.Post("/routes", handlers.CreateRoute)
	api.Get("/routes", handlers.GetRoutes)
	api.Post("/orders", handlers.CreateOrder)
	api.Get("/orders", handlers.GetOrders)

	// jwt-protected group
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET not set")
	}
	protected := api.Group("/")
	protected.Use(jwtware.New(jwtware.Config{
		SigningKey: []byte(jwtSecret),
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
		},
	}))
	protected.Post("/simulate", handlers.RunSimulation)
	protected.Get("/simulations", handlers.GetSimulations)
	protected.Get("/simulations/:id", handlers.GetSimulation)

	// seed from excel endpoint (not protected) — you can remove this in prod
	api.Post("/seed/excel", handlers.SeedFromExcel)
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello Welcome to GreenCart!")
	})
	log.Printf("Server running on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
