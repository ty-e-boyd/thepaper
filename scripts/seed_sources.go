package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/ty-e-boyd/thepaper/database"
	"github.com/ty-e-boyd/thepaper/feeds"
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

	log.Println("Seeding sources from feeds.FeedSources...")

	totalSources := 0
	totalAdded := 0
	totalSkipped := 0

	// Iterate through all categories and their feeds
	for category, feedURLs := range feeds.FeedSources {
		log.Printf("\nProcessing category: %s", category)

		for _, url := range feedURLs {
			totalSources++

			// Check if source already exists
			existing, err := database.GetSourceByURL(url)
			if err == nil && existing != nil {
				log.Printf("  ⊘ Skipped (already exists): %s", url)
				totalSkipped++
				continue
			}

			// Create new source
			// Use URL as name if we don't have a better name
			name := url
			source, err := database.CreateSource(name, category, url, true)
			if err != nil {
				log.Printf("  ✗ Failed to create source %s: %v", url, err)
				continue
			}

			log.Printf("  ✓ Added: %s (ID: %d)", url, source.ID)
			totalAdded++
		}
	}

	log.Printf("\n============================================================")
	log.Printf("Seeding complete!")
	log.Printf("Total sources processed: %d", totalSources)
	log.Printf("New sources added: %d", totalAdded)
	log.Printf("Sources skipped (already exist): %d", totalSkipped)
	log.Printf("============================================================")
}
