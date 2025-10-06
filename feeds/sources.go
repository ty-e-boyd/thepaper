package feeds

// FeedSources organizes RSS feeds by category
var FeedSources = map[string][]string{
	"General Tech News": {
		"http://feeds.feedburner.com/TechCrunch/",
		"https://techcrunch.com/feed/",
		"https://www.theverge.com/rss/index.xml",
		"http://feeds.arstechnica.com/arstechnica/index/",
		"https://www.wired.com/feed/rss",
		"http://feeds.mashable.com/Mashable",
		"https://www.engadget.com/rss.xml",
		"https://venturebeat.com/feed/",
		"https://gizmodo.com/rss",
	},

	"Hacker News": {
		"https://news.ycombinator.com/rss",
		"https://hnrss.org/newest",
		"https://hnrss.org/frontpage",
		"https://hnrss.org/ask",
		"https://hnrss.org/show",
	},

	"Programming & Development": {
		"https://dev.to/feed",
		"https://freecodecamp.org/news/rss",
		"https://stackoverflow.blog/feed",
		"https://alistapart.com/main/feed",
	},

	"Web Development & Frontend": {
		"https://smashingmagazine.com/feed",
		"https://css-tricks.com/feed",
		"https://tympanus.net/codrops/feed",
	},

	"Software Engineering & Architecture": {
		"https://martinfowler.com/feed.atom",
		"https://feeds.dzone.com/home",
	},

	"FAANG & Major Tech Companies": {
		"http://techblog.netflix.com/feeds/posts/default",
		"https://eng.uber.com/feed/",
		"https://medium.com/feed/airbnb-engineering",
		"https://stripe.com/blog/feed.rss",
		"https://engineering.fb.com/feed/",
		"https://engineering.linkedin.com/blog.rss.html",
		"http://labs.spotify.com/feed/",
		"https://github.blog/feed/",
		"https://blog.twitter.com/engineering/en_us/blog.rss",
		"http://feeds.feedburner.com/GDBcode",
	},

	"Other Tech Companies": {
		"https://dropbox.tech/feed",
		"https://instagram-engineering.com/feed",
		"https://blog.cloudflare.com/rss/",
		"https://medium.com/feed/@Pinterest_Engineering",
		"https://blog.asana.com/category/eng/feed",
	},

	"Reddit Programming": {
		"https://www.reddit.com/r/programming/.rss",
		"https://www.reddit.com/r/webdev/.rss",
		"https://www.reddit.com/r/javascript/.rss",
		"https://www.reddit.com/r/python/.rss",
		"https://www.reddit.com/r/rust/.rss",
		"https://www.reddit.com/r/golang/.rss",
	},
}

// GetAllFeeds returns a flat list of all feed URLs across all categories
func GetAllFeeds() []string {
	var allFeeds []string
	for _, feeds := range FeedSources {
		allFeeds = append(allFeeds, feeds...)
	}
	return allFeeds
}

// GetFeedsByCategory returns feeds for a specific category
func GetFeedsByCategory(category string) []string {
	return FeedSources[category]
}

// GetCategories returns all available category names
func GetCategories() []string {
	categories := make([]string, 0, len(FeedSources))
	for category := range FeedSources {
		categories = append(categories, category)
	}
	return categories
}
