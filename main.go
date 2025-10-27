package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/ty-e-boyd/thepaper/ai"
	"github.com/ty-e-boyd/thepaper/config"
	"github.com/ty-e-boyd/thepaper/database"
	"github.com/ty-e-boyd/thepaper/email"
	"github.com/ty-e-boyd/thepaper/feeds"
	"github.com/ty-e-boyd/thepaper/models"
)

const (
	topArticlesCount = 8 // Number of top articles to include in the email
)

func main() {
	// Parse command-line flags
	dryRun := flag.Bool("dry-run", false, "Run without sending emails (preview mode)")
	flag.Parse()

	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	if *dryRun {
		log.Println("🔍 DRY RUN MODE - No emails will be sent")
	}

	ctx := context.Background()

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

	// Load configuration
	log.Println("Loading configuration...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Get subscribed users from database
	users, err := database.GetAllSubscribedUsers()
	if err != nil {
		log.Fatalf("Failed to get subscribed users: %v", err)
	}
	if len(users) == 0 {
		log.Println("No subscribed users found, exiting")
		return
	}
	log.Printf("Found %d subscribed user(s)", len(users))

	// Fetch articles from RSS feeds (now pulls from database)
	feedURLs := feeds.GetAllFeeds()
	log.Printf("Fetching articles from %d feeds from database across %d categories...", len(feedURLs), len(feeds.GetCategories()))
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

	// Filter out articles sent in the last 30 days
	recentArticleURLs, err := database.GetRecentArticleURLs(30)
	if err != nil {
		log.Printf("Warning: Failed to get recent article URLs: %v", err)
		recentArticleURLs = make(map[string]bool)
	}

	var newArticles []models.Article
	for _, article := range articles {
		if !recentArticleURLs[article.Link] {
			newArticles = append(newArticles, article)
		}
	}

	duplicatesFiltered := len(articles) - len(newArticles)
	if duplicatesFiltered > 0 {
		log.Printf("Filtered out %d duplicate articles sent in the last 30 days", duplicatesFiltered)
	}
	articles = newArticles

	if len(articles) == 0 {
		log.Println("No new articles found (all were sent recently), exiting")
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

	// Count unique sources from all fetched articles
	uniqueSources := make(map[string]bool)
	for _, article := range articles {
		uniqueSources[article.Source] = true
	}

	// Create email record in database
	subject := fmt.Sprintf("The Paper - %s", time.Now().Format("January 2, 2006"))
	emailRecord, err := database.CreateEmailSent(
		subject,
		len(articles),
		len(uniqueSources),
		len(users),
	)
	if err != nil {
		log.Fatalf("Failed to create email record: %v", err)
	}
	log.Printf("✓ Email record created (ID: %d)", emailRecord.ID)

	// Save selected articles to database
	for i, article := range selectedArticles {
		_, err := database.CreateEmailArticle(
			emailRecord.ID,
			article.Link,
			article.Title,
			article.Source,
			article.RelevanceScore,
			article.Category,
			article.Tags,
			article.Summary,
			article.Published,
			i+1, // position (1-indexed)
		)
		if err != nil {
			log.Printf("Warning: Failed to save article to database: %v", err)
		}
	}
	log.Printf("✓ Saved %d articles to database", len(selectedArticles))

	// Dry run mode - skip sending
	if *dryRun {
		log.Println("\n============================================================")
		log.Println("🔍 DRY RUN SUMMARY")
		log.Println("============================================================")
		log.Printf("📊 Total articles fetched: %d", len(articles))
		log.Printf("📰 Unique sources: %d", len(uniqueSources))
		log.Printf("👥 Subscribed users: %d", len(users))
		log.Printf("⭐ Top articles selected: %d", len(selectedArticles))
		log.Println("\n📧 Would send to:")
		for _, user := range users {
			log.Printf("  • %s (%s)", user.Email, user.Name)
		}
		log.Println("\n📝 Selected articles:")
		for i, article := range selectedArticles {
			log.Printf("  %d. [%.1f] %s", i+1, article.RelevanceScore, article.Title)
			log.Printf("     Source: %s | Category: %s", article.Source, article.Category)
		}
		log.Println("\n============================================================")
		log.Println("✅ Dry run complete - no emails sent")
		log.Println("============================================================")
		return
	}

	// Send email to all subscribed users
	sender := email.NewSender(cfg.SendGridAPIKey)
	successCount := 0
	failCount := 0

	for _, user := range users {
		log.Printf("Sending email to %s (%s)...", user.Email, user.Name)

		// Build personalized HTML email
		htmlContent := email.BuildHTML(selectedArticles, len(articles), len(uniqueSources))

		err = sender.Send(cfg.FromEmail, user.Email, subject, htmlContent)
		if err != nil {
			log.Printf("  ✗ Failed to send to %s: %v", user.Email, err)
			failCount++
			continue
		}

		// Record that email was sent to this user
		_, err = database.CreateUserEmail(user.ID, emailRecord.ID)
		if err != nil {
			log.Printf("  Warning: Failed to record email send for %s: %v", user.Email, err)
		}

		log.Printf("  ✓ Sent successfully to %s", user.Email)
		successCount++
	}

	log.Println("\n============================================================")
	log.Printf("Email campaign complete!")
	log.Printf("Successfully sent: %d", successCount)
	log.Printf("Failed: %d", failCount)
	log.Printf("Total recipients: %d", len(users))
	log.Printf("Articles featured: %d", len(selectedArticles))
	log.Printf("============================================================")
}
