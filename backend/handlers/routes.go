package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/parshwanath-p2493/GreenCartt/config"
	"github.com/parshwanath-p2493/GreenCartt/models"

	"go.mongodb.org/mongo-driver/bson"
)

func CreateRoute(c *fiber.Ctx) error {
	var r models.Route
	if err := c.BodyParser(&r); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body", "details": err.Error()})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := config.DB.Collection("routes").InsertOne(ctx, r)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "insert failed", "details": err.Error()})
	}
	return c.Status(201).JSON(r)
}

func GetRoutes(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cur, err := config.DB.Collection("routes").Find(ctx, bson.M{})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "find failed", "details": err.Error()})
	}
	var routes []models.Route
	if err := cur.All(ctx, &routes); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "decode failed", "details": err.Error()})
	}
	return c.JSON(routes)
}

func GetRoute(c *fiber.Ctx) error {
	id := c.Params("id") // here id is route_id string
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var r models.Route
	if err := config.DB.Collection("routes").FindOne(ctx, bson.M{"route_id": id}).Decode(&r); err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}
	return c.JSON(r)
}

func UpdateRoute(c *fiber.Ctx) error {
	id := c.Params("id")
	var r models.Route
	if err := c.BodyParser(&r); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := config.DB.Collection("routes").UpdateOne(ctx, bson.M{"route_id": id}, bson.M{"$set": r})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "update failed", "details": err.Error()})
	}
	return c.JSON(r)
}

func DeleteRoute(c *fiber.Ctx) error {
	id := c.Params("id")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := config.DB.Collection("routes").DeleteOne(ctx, bson.M{"route_id": id})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "delete failed", "details": err.Error()})
	}
	return c.SendStatus(204)
}
