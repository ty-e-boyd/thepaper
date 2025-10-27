package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/ty-e-boyd/thepaper/database"
)

func main() {
	// Load .env file
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Connect to database
	log.Println("Connecting to database...")
	if err := database.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Run migrations
	if err := database.AutoMigrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	email := "tyler@tylerevan.dev"
	name := "Tyler"

	log.Printf("Adding user: %s (%s)", email, name)

	// Check if user already exists
	existing, err := database.GetUserByEmail(email)
	if err == nil && existing != nil {
		log.Printf("✓ User already exists (ID: %d)", existing.ID)
		log.Printf("  Email: %s", existing.Email)
		log.Printf("  Name: %s", existing.Name)
		log.Printf("  Subscribed: %v", existing.Subscribed)
		log.Printf("  Unsubscribe Token: %s", existing.UnsubscribeToken)
		return
	}

	// Create new user
	user, err := database.CreateUser(email, name)
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

	log.Printf("✓ User created successfully!")
	log.Printf("  ID: %d", user.ID)
	log.Printf("  Email: %s", user.Email)
	log.Printf("  Name: %s", user.Name)
	log.Printf("  Subscribed: %v", user.Subscribed)
	log.Printf("  Unsubscribe Token: %s", user.UnsubscribeToken)
	log.Printf("  Created At: %s", user.CreatedAt)
}
