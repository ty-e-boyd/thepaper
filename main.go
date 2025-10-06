package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/ty-e-boyd/thepaper/ai"
	"github.com/ty-e-boyd/thepaper/config"
	"github.com/ty-e-boyd/thepaper/email"
	"github.com/ty-e-boyd/thepaper/feeds"
)

const (
	topArticlesCount = 5 // Number of top articles to include in the email
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	ctx := context.Background()

	// Load configuration
	log.Println("Loading configuration...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Fetch articles from RSS feeds
	feedURLs := feeds.GetAllFeeds()
	log.Printf("Fetching articles from %d feeds across %d categories...", len(feedURLs), len(feeds.GetCategories()))
	fetcher := feeds.NewFetcher()
	articles, err := fetcher.FetchAll(feedURLs)
	if err != nil {
		log.Fatalf("Failed to fetch articles: %v", err)
	}
	log.Printf("Fetched %d articles", len(articles))

	if len(articles) == 0 {
		log.Println("No articles found, exiting")
		return
	}

	// Show article distribution by source
	sourceCount := make(map[string]int)
	for _, article := range articles {
		sourceCount[article.Source]++
	}
	log.Printf("\nArticle distribution by source:")
	for source, count := range sourceCount {
		log.Printf("  %s: %d articles", source, count)
	}
	log.Println()

	// Analyze and select top articles using Gemini
	log.Println("Analyzing articles with Gemini AI...")
	analyzer, err := ai.NewAnalyzer(ctx, cfg.GeminiAPIKey)
	if err != nil {
		log.Fatalf("Failed to create analyzer: %v", err)
	}
	defer analyzer.Close()

	selectedArticles, err := analyzer.SelectAndSummarize(ctx, articles, topArticlesCount)
	if err != nil {
		log.Fatalf("Failed to analyze articles: %v", err)
	}
	log.Printf("Selected and summarized %d top articles", len(selectedArticles))

	// Build HTML email
	log.Println("Building email...")
	htmlContent := email.BuildHTML(selectedArticles)

	// Send email via SendGrid
	log.Println("Sending email via SendGrid...")
	sender := email.NewSender(cfg.SendGridAPIKey)
	subject := fmt.Sprintf("The Paper - %s", time.Now().Format("January 2, 2006"))

	err = sender.Send(cfg.FromEmail, cfg.ToEmail, subject, htmlContent)
	if err != nil {
		log.Fatalf("Failed to send email: %v", err)
	}

	log.Println(" Email sent successfully!")
}
