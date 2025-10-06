# The Paper

A daily programming news aggregator that fetches articles from RSS feeds, uses Gemini AI to select the most relevant content, and sends a curated HTML email digest via SendGrid.

## Features

- Fetches articles from multiple RSS feeds concurrently
- Uses Gemini AI to score articles for relevance
- Generates AI-powered summaries for selected articles
- Sends beautifully formatted HTML emails via SendGrid
- Configurable via environment variables

## Setup

1. Install dependencies:
```bash
go mod download
```

2. Copy `.env.example` to `.env` and configure:
```bash
cp .env.example .env
```

3. Set required environment variables in `.env`:
   - `GEMINI_API_KEY`: Your Google Gemini API key
   - `SENDGRID_API_KEY`: Your SendGrid API key
   - `FROM_EMAIL`: Sender email address
   - `TO_EMAIL`: Recipient email address

## Usage

Run the script:
```bash
go run main.go
```

Or build and run:
```bash
go build
./thepaper
```

## Configuration

The app selects the top 5 articles by default. Modify `topArticlesCount` in `main.go` to change this.

## RSS Feeds

Feed sources are configured in `feeds/sources.go` and organized by category:
- General Tech News (TechCrunch, The Verge, Ars Technica, etc.)
- Hacker News (frontpage, show HN, ask HN, etc.)
- Programming & Development (DEV.to, freeCodeCamp, Stack Overflow)
- Web Development & Frontend (Smashing Magazine, CSS-Tricks)
- Software Engineering & Architecture
- FAANG & Major Tech Companies (Netflix, Uber, Airbnb, etc.)
- Other Tech Companies (Dropbox, Cloudflare, Pinterest, etc.)
- Reddit Programming (r/programming, r/webdev, r/javascript, etc.)

To modify feed sources, edit `feeds/sources.go`.

## Project Structure

```
thepaper/
├── main.go              # Entry point and orchestration
├── models/
│   └── types.go         # Data structures
├── config/
│   └── config.go        # Environment configuration
├── feeds/
│   ├── sources.go       # RSS feed URLs organized by category
│   └── fetcher.go       # RSS feed fetching
├── ai/
│   └── analyzer.go      # Gemini-powered analysis
└── email/
    ├── builder.go       # HTML email generation
    └── sender.go        # SendGrid integration
```
