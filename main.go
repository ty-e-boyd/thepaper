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
	"github.com/ty-e-boyd/thepaper/models"
)

const (
	topArticlesCount = 8 // Number of top articles to include in the email
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

	// Filter articles to last 24 hours
	cutoff := time.Now().Add(-24 * time.Hour)
	var recentArticles []models.Article
	for _, article := range articles {
		if article.Published.After(cutoff) || article.Published.IsZero() {
			recentArticles = append(recentArticles, article)
		}
	}
	log.Printf("Filtered to %d articles from last 24 hours (from %d total)\n", len(recentArticles), len(articles))
	articles = recentArticles

	if len(articles) == 0 {
		log.Println("No recent articles found, exiting")
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
	log.Printf("Analyzing articles with Gemini AI (rate limit: %v)...\n", cfg.GeminiRateLimit)
	analyzer, err := ai.NewAnalyzer(ctx, cfg.GeminiAPIKey, cfg.GeminiRateLimit)
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

	// Count unique sources from all fetched articles
	uniqueSources := make(map[string]bool)
	for _, article := range articles {
		uniqueSources[article.Source] = true
	}

	htmlContent := email.BuildHTML(selectedArticles, len(articles), len(uniqueSources))

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
