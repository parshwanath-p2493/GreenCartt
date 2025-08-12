package handlers

import (
	"context"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"

	"github.com/parshwanath-p2493/GreenCartt/config"
)

type Manager struct {
	Email    string `bson:"email" json:"email"`
	Password string `bson:"password,omitempty" json:"-"`
}

func SeedManager(c *fiber.Ctx) error {
	// only allow if env seed vars set
	email := os.Getenv("SEED_MANAGER_EMAIL")
	pass := os.Getenv("SEED_MANAGER_PASSWORD")
	if email == "" || pass == "" {
		return c.Status(400).JSON(fiber.Map{"error": "seed env not configured"})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// check exists
	count, _ := config.DB.Collection("managers").CountDocuments(ctx, bson.M{"email": email})
	if count > 0 {
		return c.JSON(fiber.Map{"status": "already seeded"})
	}
	hashed, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	m := Manager{Email: email, Password: string(hashed)}
	_, err := config.DB.Collection("managers").InsertOne(ctx, m)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "insert failed", "details": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "seeded"})
}
func Register(c *fiber.Ctx) error {
	// var name string
	// var email string
	// var password string
	name := "Parshwanath"
	email := "manager@example.com"
	password := "Password@123"
	if err := c.BodyParser(&fiber.Map{
		"name":     &name,
		"email":    &email,
		"password": &password,
	}); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// check exists
	count, _ := config.DB.Collection("managers").CountDocuments(ctx, bson.M{"email": email})
	if count > 0 {
		return c.JSON(fiber.Map{"status": "already exists"})
	}
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	m := Manager{Email: email, Password: string(hashed)}
	_, err := config.DB.Collection("managers").InsertOne(ctx, m)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "insert failed", "details": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "registered"})
}
func Login(c *fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var m Manager
	if err := config.DB.Collection("managers").FindOne(ctx, bson.M{"email": body.Email}).Decode(&m); err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "invalid credentials"})
	}
	if err := bcrypt.CompareHashAndPassword([]byte(m.Password), []byte(body.Password)); err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "invalid credentials"})
	}
	// create jwt
	secret := os.Getenv("JWT_SECRET")
	claims := jwt.MapClaims{
		"email": m.Email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(secret))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "token error"})
	}
	return c.JSON(fiber.Map{"token": ss, "user": fiber.Map{"email": m.Email}})
}
