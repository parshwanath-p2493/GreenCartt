package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/parshwanath-p2493/GreenCartt/config"
	"github.com/parshwanath-p2493/GreenCartt/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetSimulations returns paginated list (simple)
func GetSimulations(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cur, err := config.DB.Collection("simulations").Find(ctx, bson.M{}, nil)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "find failed", "details": err.Error()})
	}
	var sims []models.SimulationResult
	if err := cur.All(ctx, &sims); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "decode failed", "details": err.Error()})
	}
	return c.JSON(sims)
}

// GetSimulation returns a single simulation by id
func GetSimulation(c *fiber.Ctx) error {
	id := c.Params("id")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}
	var s models.SimulationResult
	if err := config.DB.Collection("simulations").FindOne(ctx, filter).Decode(&s); err != nil {
		// try primitive.ObjectID fallback (if your saved id used ObjectID)
		oid, err2 := primitive.ObjectIDFromHex(id)
		if err2 == nil {
			if err3 := config.DB.Collection("simulations").FindOne(ctx, bson.M{"_id": oid}).Decode(&s); err3 == nil {
				return c.JSON(s)
			}
		}
		return c.Status(404).JSON(fiber.Map{"error": "not found", "details": err.Error()})
	}
	return c.JSON(s)
}
