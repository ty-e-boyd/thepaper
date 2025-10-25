package ai

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ty-e-boyd/thepaper/models"
	"google.golang.org/genai"
)

// Analyzer uses Gemini AI to select and summarize articles
type Analyzer struct {
	client          *genai.Client
	rateLimitDelay  time.Duration
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
		client:          client,
		rateLimitDelay:  rateLimitDelay,
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

	// Extract tags and categories for top candidates (check more than topN for diversity)
	candidateCount := topN * 3
	if candidateCount > len(analyzed) {
		candidateCount = len(analyzed)
	}

	log.Printf("\nExtracting tags and categories for top %d candidates...", candidateCount)
	for i := 0; i < candidateCount; i++ {
		tags, category, err := a.extractTagsAndCategory(ctx, analyzed[i].Article)
		if err != nil {
			log.Printf("  ✗ Error extracting tags for '%s': %v", analyzed[i].Title, err)
			tags = []string{}
			category = "General"
		} else {
			log.Printf("  ✓ '%s' → Category: %s, Tags: %v", analyzed[i].Title, category, tags)
		}
		analyzed[i].Tags = tags
		analyzed[i].Category = category
	}

	// Select top N articles with diversity constraints
	selected := a.selectWithDiversity(analyzed, topN)

	log.Printf("\nSelected %d articles with category diversity:", len(selected))
	for i, article := range selected {
		log.Printf("  %d. [%.1f] %s (Category: %s)", i+1, article.RelevanceScore, article.Title, article.Category)
	}

	// Summarize selected articles
	log.Printf("\nGenerating summaries...")
	for i := range selected {
		selected[i].Selected = true
		summary, err := a.summarizeArticle(ctx, selected[i].Article)
		if err != nil {
			log.Printf("  ✗ Error summarizing '%s': %v", selected[i].Title, err)
			summary = "Summary unavailable."
		} else {
			log.Printf("  ✓ Summarized '%s'", selected[i].Title)
		}
		selected[i].Summary = summary
	}

	return selected, nil
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

Respond with ONLY a number between 0 and 10. You may use half increments (e.g., 7.5, 8.5, 9.5).`, article.Title, article.Description)

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
	prompt := fmt.Sprintf(`Summarize the following article in ONE concise sentence for a technical audience.
Focus on the single most important technical point or takeaway.

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

// extractTagsAndCategory uses Gemini to extract relevant tags and categorize the article
func (a *Analyzer) extractTagsAndCategory(ctx context.Context, article models.Article) ([]string, string, error) {
	prompt := fmt.Sprintf(`Analyze this article and provide:
1. A category (ONE of: AI/ML, Web Development, Backend, DevOps, Mobile, Security, Data, Cloud, Open Source, Career, General)
2. 2-3 relevant tags (short keywords)

Article:
Title: %s
Description: %s

Respond in this EXACT format:
Category: [category]
Tags: [tag1, tag2, tag3]`, article.Title, article.Description)

	var responseText string
	err := retryWithBackoff(ctx, 5, func() error {
		// Rate limit before making request
		a.rateLimit()

		content := []*genai.Content{{Parts: []*genai.Part{genai.NewPartFromText(prompt)}}}
		response, err := a.client.Models.GenerateContent(ctx, "gemini-2.0-flash", content, nil)
		if err != nil {
			return fmt.Errorf("failed to extract tags: %w", err)
		}

		responseText = strings.TrimSpace(response.Text())
		return nil
	})

	if err != nil {
		return nil, "", err
	}

	// Parse response
	lines := strings.Split(responseText, "\n")
	var category string
	var tags []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Category:") {
			category = strings.TrimSpace(strings.TrimPrefix(line, "Category:"))
		} else if strings.HasPrefix(line, "Tags:") {
			tagStr := strings.TrimSpace(strings.TrimPrefix(line, "Tags:"))
			tagStr = strings.Trim(tagStr, "[]")
			for _, tag := range strings.Split(tagStr, ",") {
				tag = strings.TrimSpace(tag)
				if tag != "" {
					tags = append(tags, tag)
				}
			}
		}
	}

	if category == "" {
		category = "General"
	}
	if len(tags) == 0 {
		tags = []string{"tech"}
	}

	return tags, category, nil
}

// selectWithDiversity selects top N articles while ensuring category diversity and topic uniqueness
func (a *Analyzer) selectWithDiversity(analyzed []models.AnalyzedArticle, topN int) []models.AnalyzedArticle {
	selected := make([]models.AnalyzedArticle, 0, topN)
	categoryCount := make(map[string]int)
	const maxPerCategory = 2

	// Keep track of selected article topics for duplicate detection
	selectedTopics := make([]string, 0, topN)

	for _, article := range analyzed {
		if len(selected) >= topN {
			break
		}

		// Check category limit
		if categoryCount[article.Category] >= maxPerCategory {
			log.Printf("  ⊘ Skipping '%s' - category '%s' limit reached (%d/%d)",
				article.Title, article.Category, categoryCount[article.Category], maxPerCategory)
			continue
		}

		// Check for duplicate topics
		if a.isDuplicateTopic(article.Title, selectedTopics) {
			log.Printf("  ⊘ Skipping '%s' - similar topic already selected", article.Title)
			continue
		}

		// Add to selected
		selected = append(selected, article)
		categoryCount[article.Category]++
		selectedTopics = append(selectedTopics, article.Title)
	}

	return selected
}

// isDuplicateTopic checks if an article title is too similar to already selected topics
func (a *Analyzer) isDuplicateTopic(title string, selectedTopics []string) bool {
	titleLower := strings.ToLower(title)
	titleWords := strings.Fields(titleLower)

	for _, selectedTitle := range selectedTopics {
		selectedLower := strings.ToLower(selectedTitle)
		selectedWords := strings.Fields(selectedLower)

		// Count common significant words (ignore common words)
		commonWords := 0
		insignificantWords := map[string]bool{
			"the": true, "a": true, "an": true, "and": true, "or": true,
			"but": true, "in": true, "on": true, "at": true, "to": true,
			"for": true, "of": true, "with": true, "by": true, "from": true,
			"is": true, "are": true, "was": true, "were": true, "be": true,
			"how": true, "why": true, "what": true, "when": true, "where": true,
		}

		for _, word := range titleWords {
			if len(word) <= 2 || insignificantWords[word] {
				continue
			}
			for _, selectedWord := range selectedWords {
				if len(selectedWord) <= 2 || insignificantWords[selectedWord] {
					continue
				}
				// Check if words match or are very similar
				if word == selectedWord || strings.HasPrefix(word, selectedWord) || strings.HasPrefix(selectedWord, word) {
					commonWords++
					break
				}
			}
		}

		// If more than 40% of significant words match, consider it a duplicate
		significantTitleWords := 0
		for _, word := range titleWords {
			if len(word) > 2 && !insignificantWords[word] {
				significantTitleWords++
			}
		}

		if significantTitleWords > 0 && float64(commonWords)/float64(significantTitleWords) > 0.4 {
			return true
		}
	}

	return false
}
