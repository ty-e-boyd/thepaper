package feeds

import (
	"fmt"
	"log"
	"sync"

	"github.com/mmcdole/gofeed"
	"github.com/ty-e-boyd/thepaper/models"
)

// Fetcher handles fetching and parsing RSS feeds
type Fetcher struct {
	parser *gofeed.Parser
}

// NewFetcher creates a new RSS feed fetcher
func NewFetcher() *Fetcher {
	return &Fetcher{
		parser: gofeed.NewParser(),
	}
}

// FetchAll fetches articles from all provided feed URLs concurrently
func (f *Fetcher) FetchAll(feedURLs []string) ([]models.Article, error) {
	var wg sync.WaitGroup
	articlesChan := make(chan []models.Article, len(feedURLs))
	errorsChan := make(chan error, len(feedURLs))

	for _, url := range feedURLs {
		wg.Add(1)
		go func(feedURL string) {
			defer wg.Done()

			articles, err := f.fetchSingle(feedURL)
			if err != nil {
				errorsChan <- fmt.Errorf("error fetching %s: %w", feedURL, err)
				return
			}
			articlesChan <- articles
		}(url)
	}

	wg.Wait()
	close(articlesChan)
	close(errorsChan)

	// Collect all articles
	var allArticles []models.Article
	for articles := range articlesChan {
		allArticles = append(allArticles, articles...)
	}

	// Check for errors (non-fatal, just log them)
	var errors []error
	for err := range errorsChan {
		errors = append(errors, err)
	}

	successCount := len(feedURLs) - len(errors)
	log.Printf("\nFetch summary: %d/%d feeds successful, %d failed", successCount, len(feedURLs), len(errors))

	if len(errors) > 0 && len(allArticles) == 0 {
		return nil, fmt.Errorf("all feeds failed: %v", errors)
	}

	// Deduplicate by URL
	beforeDedup := len(allArticles)
	allArticles = deduplicateArticles(allArticles)
	log.Printf("Deduplication: %d articles → %d unique articles\n", beforeDedup, len(allArticles))

	return allArticles, nil
}

// fetchSingle fetches and parses a single RSS feed
func (f *Fetcher) fetchSingle(feedURL string) ([]models.Article, error) {
	feed, err := f.parser.ParseURL(feedURL)
	if err != nil {
		log.Printf("  ✗ Failed to fetch %s: %v", feedURL, err)
		return nil, err
	}

	articles := make([]models.Article, 0, len(feed.Items))
	for _, item := range feed.Items {
		article := models.Article{
			Title:       item.Title,
			Description: item.Description,
			Link:        item.Link,
			Source:      feed.Title,
		}

		// Set published time
		if item.PublishedParsed != nil {
			article.Published = *item.PublishedParsed
		}

		// Use content if available, otherwise use description
		if item.Content != "" {
			article.Content = item.Content
		} else {
			article.Content = item.Description
		}

		articles = append(articles, article)
	}

	log.Printf("  ✓ Fetched %d articles from %s", len(articles), feed.Title)
	return articles, nil
}

// deduplicateArticles removes duplicate articles based on URL
func deduplicateArticles(articles []models.Article) []models.Article {
	seen := make(map[string]bool)
	unique := make([]models.Article, 0, len(articles))

	for _, article := range articles {
		if !seen[article.Link] {
			seen[article.Link] = true
			unique = append(unique, article)
		}
	}

	return unique
}
