package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/parshwanath-p2493/GreenCartt/config"
	"github.com/parshwanath-p2493/GreenCartt/models"

	"go.mongodb.org/mongo-driver/bson"
)

func CreateOrder(c *fiber.Ctx) error {
	var o models.Order
	if err := c.BodyParser(&o); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body", "details": err.Error()})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := config.DB.Collection("orders").InsertOne(ctx, o)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "insert failed", "details": err.Error()})
	}
	return c.Status(201).JSON(o)
}

func GetOrders(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cur, err := config.DB.Collection("orders").Find(ctx, bson.M{})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "find failed", "details": err.Error()})
	}
	var orders []models.Order
	if err := cur.All(ctx, &orders); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "decode failed", "details": err.Error()})
	}
	return c.JSON(orders)
}

func GetOrder(c *fiber.Ctx) error {
	id := c.Params("id")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var o models.Order
	if err := config.DB.Collection("orders").FindOne(ctx, bson.M{"order_id": id}).Decode(&o); err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}
	return c.JSON(o)
}

func UpdateOrder(c *fiber.Ctx) error {
	id := c.Params("id")
	var o models.Order
	if err := c.BodyParser(&o); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := config.DB.Collection("orders").UpdateOne(ctx, bson.M{"order_id": id}, bson.M{"$set": o})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "update failed", "details": err.Error()})
	}
	return c.JSON(o)
}

func DeleteOrder(c *fiber.Ctx) error {
	id := c.Params("id")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := config.DB.Collection("orders").DeleteOne(ctx, bson.M{"order_id": id})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "delete failed", "details": err.Error()})
	}
	return c.SendStatus(204)
}
