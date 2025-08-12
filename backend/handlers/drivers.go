package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/parshwanath-p2493/GreenCartt/config"
	"github.com/parshwanath-p2493/GreenCartt/models"
	"go.mongodb.org/mongo-driver/bson"
)

func CreateDriver(c *fiber.Ctx) error {
	var d models.Driver
	if err := c.BodyParser(&d); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body", "details": err.Error()})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := config.DB.Collection("drivers").InsertOne(ctx, d)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "db insert failed", "details": err.Error()})
	}
	return c.Status(201).JSON(d)
}

func GetDrivers(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cur, err := config.DB.Collection("drivers").Find(ctx, bson.M{})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "db find failed", "details": err.Error()})
	}
	var drivers []models.Driver
	if err := cur.All(ctx, &drivers); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "db decode failed", "details": err.Error()})
	}
	return c.JSON(drivers)
}

func GetDriver(c *fiber.Ctx) error {
	id := c.Params("id")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var d models.Driver
	if err := config.DB.Collection("drivers").FindOne(ctx, bson.M{"id": id}).Decode(&d); err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}
	return c.JSON(d)
}

func UpdateDriver(c *fiber.Ctx) error {
	id := c.Params("id")
	var d models.Driver
	if err := c.BodyParser(&d); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body", "details": err.Error()})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := config.DB.Collection("drivers").UpdateOne(ctx, bson.M{"id": id}, bson.M{"$set": d})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "update failed", "details": err.Error()})
	}
	return c.JSON(d)
}

func DeleteDriver(c *fiber.Ctx) error {
	id := c.Params("id")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := config.DB.Collection("drivers").DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "delete failed", "details": err.Error()})
	}
	return c.SendStatus(204)
}
