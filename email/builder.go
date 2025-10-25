package email

import (
	"fmt"
	"strings"
	"time"

	"github.com/ty-e-boyd/thepaper/models"
)

// capitalizeTag capitalizes the first letter of each word in a tag
func capitalizeTag(tag string) string {
	if tag == "" {
		return tag
	}

	words := strings.Fields(tag)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, " ")
}

// getTagColor returns a color for a tag based on its content
func getTagColor(tag string) string {
	colors := []string{
		"#3498db", // blue
		"#e74c3c", // red
		"#2ecc71", // green
		"#f39c12", // orange
		"#9b59b6", // purple
		"#1abc9c", // turquoise
		"#e67e22", // carrot
		"#34495e", // dark gray
		"#16a085", // green sea
		"#c0392b", // dark red
		"#8e44ad", // wisteria
		"#27ae60", // nephritis
	}

	// Simple hash function to assign consistent colors to tags
	hash := 0
	for _, char := range tag {
		hash += int(char)
	}

	return colors[hash%len(colors)]
}

// BuildHTML generates an HTML email from analyzed articles
func BuildHTML(articles []models.AnalyzedArticle, totalArticles, totalSources int) string {
	var sb strings.Builder

	// Email header and styles
	sb.WriteString(`<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<style>
		body {
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
			line-height: 1.6;
			color: #333;
			max-width: 600px;
			margin: 0 auto;
			padding: 20px;
			background-color: #f5f5f5;
		}
		.container {
			background-color: #ffffff;
			padding: 30px;
			border-radius: 8px;
			box-shadow: 0 2px 4px rgba(0,0,0,0.1);
		}
		h1 {
			color: #2c3e50;
			font-size: 28px;
			margin-bottom: 10px;
			border-bottom: 3px solid #3498db;
			padding-bottom: 10px;
		}
		.date {
			color: #7f8c8d;
			font-size: 14px;
			margin-bottom: 30px;
		}
		.article {
			margin-bottom: 30px;
			padding-bottom: 20px;
			border-bottom: 1px solid #ecf0f1;
		}
		.article:last-child {
			border-bottom: none;
		}
		.article-title {
			font-size: 20px;
			font-weight: 600;
			color: #2c3e50;
			margin-bottom: 8px;
		}
		.article-title a {
			color: #2c3e50;
			text-decoration: none;
		}
		.article-title a:hover {
			color: #3498db;
		}
		.article-meta {
			font-size: 13px;
			color: #7f8c8d;
			margin-bottom: 8px;
		}
		.article-tags {
			margin-bottom: 12px;
		}
		.tag {
			display: inline-block;
			color: white;
			padding: 3px 10px;
			border-radius: 12px;
			font-size: 11px;
			font-weight: 500;
			margin-right: 6px;
			margin-bottom: 4px;
		}
		.category-badge {
			display: inline-block;
			background-color: #3498db;
			color: white;
			padding: 3px 10px;
			border-radius: 12px;
			font-size: 11px;
			font-weight: 600;
			margin-right: 8px;
		}
		.article-summary {
			color: #555;
			line-height: 1.7;
			margin-bottom: 10px;
		}
		.read-more {
			display: inline-block;
			color: #3498db;
			text-decoration: none;
			font-weight: 500;
			font-size: 14px;
		}
		.read-more:hover {
			text-decoration: underline;
		}
		.stats {
			margin-top: 30px;
			padding: 15px;
			background-color: #f8f9fa;
			border-radius: 6px;
			text-align: center;
			font-size: 13px;
			color: #555;
		}
		.stats strong {
			color: #2c3e50;
		}
		.footer {
			margin-top: 20px;
			padding-top: 20px;
			border-top: 2px solid #ecf0f1;
			text-align: center;
			color: #95a5a6;
			font-size: 12px;
		}
	</style>
</head>
<body>
	<div class="container">
		<h1>📰 The Paper</h1>
		<div class="date">` + time.Now().Format("Monday, January 2, 2006") + `</div>
`)

	// Add articles
	for i, article := range articles {
		// Build tags HTML
		tagsHTML := ""
		if len(article.Tags) > 0 {
			for _, tag := range article.Tags {
				capitalizedTag := capitalizeTag(tag)
				color := getTagColor(tag)
				tagsHTML += fmt.Sprintf(`<span class="tag" style="background-color: %s;">%s</span>`, color, escapeHTML(capitalizedTag))
			}
		}

		sb.WriteString(fmt.Sprintf(`
		<div class="article">
			<div class="article-title">
				<a href="%s" target="_blank">%d. %s</a>
			</div>
			<div class="article-meta">
				<span class="category-badge">%s</span>
				Source: %s | Score: %.1f/10
			</div>
			<div class="article-tags">
				%s
			</div>
			<div class="article-summary">
				%s
			</div>
			<a href="%s" class="read-more" target="_blank">Read full article →</a>
		</div>
`, article.Link, i+1, escapeHTML(article.Title), escapeHTML(article.Category),
			escapeHTML(article.Source), article.RelevanceScore, tagsHTML,
			escapeHTML(article.Summary), article.Link))
	}

	// Stats section
	sb.WriteString(fmt.Sprintf(`
		<div class="stats">
			<p>📊 <strong>Today's Digest Stats:</strong> Analyzed <strong>%d articles</strong> from <strong>%d sources</strong></p>
		</div>
`, totalArticles, totalSources))

	// Footer
	sb.WriteString(`
		<div class="footer">
			<p>You're receiving this because you subscribed to The Paper daily digest.</p>
			<p>Curated and summarized by AI | Powered by Gemini</p>
		</div>
	</div>
</body>
</html>`)

	return sb.String()
}

// escapeHTML escapes special HTML characters
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}
