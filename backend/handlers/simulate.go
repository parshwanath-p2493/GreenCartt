package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/parshwanath-p2493/GreenCartt/config"
	"github.com/parshwanath-p2493/GreenCartt/models"
	"github.com/parshwanath-p2493/GreenCartt/services"
	"go.mongodb.org/mongo-driver/bson"
)

func RunSimulation(c *fiber.Ctx) error {
	var input models.SimulationInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body", "details": err.Error()})
	}

	// load drivers, routes, orders from DB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	driverCur, err := config.DB.Collection("drivers").Find(ctx, bson.M{})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "db error", "details": err.Error()})
	}
	var drivers []models.Driver
	if err := driverCur.All(ctx, &drivers); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "db decode error", "details": err.Error()})
	}

	routeCur, err := config.DB.Collection("routes").Find(ctx, bson.M{})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "db error", "details": err.Error()})
	}
	var routes []models.Route
	if err := routeCur.All(ctx, &routes); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "db decode", "details": err.Error()})
	}

	orderCur, err := config.DB.Collection("orders").Find(ctx, bson.M{})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "db error", "details": err.Error()})
	}
	var orders []models.Order
	if err := orderCur.All(ctx, &orders); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "db decode", "details": err.Error()})
	}

	// validate input
	if input.AvailableDrivers <= 0 || input.AvailableDrivers > len(drivers) {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid parameter", "details": fiber.Map{"available_drivers": "must be >0 and <= total drivers"}})
	}

	// call simulate service
	result, err := services.Simulate(drivers, routes, orders, input)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "simulation error", "details": err.Error()})
	}
	// store result
	result.ID = uuid.New().String()
	result.Timestamp = time.Now().UTC()

	_, err = config.DB.Collection("simulations").InsertOne(ctx, result)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "db insert failed", "details": err.Error()})
	}

	return c.Status(200).JSON(result)
}
