package ai

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"google.golang.org/genai"
	"github.com/ty-e-boyd/thepaper/models"
)

// Analyzer uses Gemini AI to select and summarize articles
type Analyzer struct {
	client        *genai.Client
	rateLimitDelay time.Duration
	lastRequestTime time.Time
}

// NewAnalyzer creates a new Gemini-powered analyzer with rate limiting
func NewAnalyzer(ctx context.Context, apiKey string, rateLimitDelay time.Duration) (*Analyzer, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &Analyzer{
		client:         client,
		rateLimitDelay: rateLimitDelay,
		lastRequestTime: time.Now(),
	}, nil
}

// Close cleans up the analyzer resources
func (a *Analyzer) Close() {
	// Client cleanup if needed
}

// rateLimit ensures we don't exceed API rate limits
func (a *Analyzer) rateLimit() {
	elapsed := time.Since(a.lastRequestTime)
	if elapsed < a.rateLimitDelay {
		time.Sleep(a.rateLimitDelay - elapsed)
	}
	a.lastRequestTime = time.Now()
}

// retryWithBackoff retries a function with exponential backoff on rate limit errors
func retryWithBackoff(ctx context.Context, maxRetries int, fn func() error) error {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s, 8s, 16s
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			log.Printf("  Retry attempt %d/%d after %v...", attempt, maxRetries, backoff)
			time.Sleep(backoff)
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err
		// Check if it's a rate limit error
		if !strings.Contains(err.Error(), "RESOURCE_EXHAUSTED") && !strings.Contains(err.Error(), "429") {
			// Not a rate limit error, don't retry
			return err
		}
	}
	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// SelectAndSummarize analyzes articles, scores them for relevance, and summarizes the top ones
func (a *Analyzer) SelectAndSummarize(ctx context.Context, articles []models.Article, topN int) ([]models.AnalyzedArticle, error) {
	if len(articles) == 0 {
		return nil, fmt.Errorf("no articles to analyze")
	}

	// Score all articles for relevance
	log.Printf("Scoring %d articles...", len(articles))
	analyzed := make([]models.AnalyzedArticle, len(articles))
	for i, article := range articles {
		score, err := a.scoreArticle(ctx, article)
		if err != nil {
			log.Printf("  ✗ Error scoring '%s' from %s: %v", article.Title, article.Source, err)
			score = 0
		} else {
			log.Printf("  %.1f - %s (from %s)", score, article.Title, article.Source)
		}

		analyzed[i] = models.AnalyzedArticle{
			Article:        article,
			RelevanceScore: score,
			Selected:       false,
		}
	}

	// Sort by relevance score
	sort.Slice(analyzed, func(i, j int) bool {
		return analyzed[i].RelevanceScore > analyzed[j].RelevanceScore
	})

	// Select top N articles
	if topN > len(analyzed) {
		topN = len(analyzed)
	}

	log.Printf("\nTop %d articles selected:", topN)
	for i := 0; i < topN; i++ {
		log.Printf("  %d. [%.1f] %s", i+1, analyzed[i].RelevanceScore, analyzed[i].Title)
	}

	// Summarize selected articles
	log.Printf("\nGenerating summaries...")
	for i := 0; i < topN; i++ {
		analyzed[i].Selected = true
		summary, err := a.summarizeArticle(ctx, analyzed[i].Article)
		if err != nil {
			log.Printf("  ✗ Error summarizing '%s': %v", analyzed[i].Title, err)
			summary = "Summary unavailable."
		} else {
			log.Printf("  ✓ Summarized '%s'", analyzed[i].Title)
		}
		analyzed[i].Summary = summary
	}

	return analyzed[:topN], nil
}

// scoreArticle uses Gemini to score an article's relevance for programming/tech news
func (a *Analyzer) scoreArticle(ctx context.Context, article models.Article) (float64, error) {
	prompt := fmt.Sprintf(`Rate the following article's relevance for a daily programming and technology newsletter on a scale of 0-10.
Consider:
- Technical depth and value
- Relevance to software developers
- Timeliness and importance
- Novelty and interest

Article:
Title: %s
Description: %s

Respond with ONLY a number between 0 and 10.`, article.Title, article.Description)

	var score float64
	err := retryWithBackoff(ctx, 5, func() error {
		// Rate limit before making request
		a.rateLimit()

		content := []*genai.Content{{Parts: []*genai.Part{genai.NewPartFromText(prompt)}}}
		response, err := a.client.Models.GenerateContent(ctx, "gemini-2.0-flash", content, nil)
		if err != nil {
			return fmt.Errorf("failed to score article: %w", err)
		}

		// Extract score from response
		scoreStr := strings.TrimSpace(response.Text())
		parsedScore, err := strconv.ParseFloat(scoreStr, 64)
		if err != nil {
			return fmt.Errorf("invalid score format: %s", scoreStr)
		}
		score = parsedScore
		return nil
	})

	if err != nil {
		return 0, err
	}
	return score, nil
}

// summarizeArticle uses Gemini to create a concise summary
func (a *Analyzer) summarizeArticle(ctx context.Context, article models.Article) (string, error) {
	prompt := fmt.Sprintf(`Summarize the following article in 2-3 concise sentences for a technical audience.
Focus on the key technical points and why it matters.

Article:
Title: %s
Content: %s

Summary:`, article.Title, article.Content)

	var summary string
	err := retryWithBackoff(ctx, 5, func() error {
		// Rate limit before making request
		a.rateLimit()

		content := []*genai.Content{{Parts: []*genai.Part{genai.NewPartFromText(prompt)}}}
		response, err := a.client.Models.GenerateContent(ctx, "gemini-2.0-flash", content, nil)
		if err != nil {
			return fmt.Errorf("failed to summarize article: %w", err)
		}

		summary = strings.TrimSpace(response.Text())
		return nil
	})

	if err != nil {
		return "", err
	}
	return summary, nil
}
