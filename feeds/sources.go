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
		"https://hashnode.com/rss",
		"https://www.infoq.com/feed",
		"https://thenewstack.io/feed",
		"https://changelog.com/posts/feed",
		"https://daily.dev/blog/rss.xml",
	},

	"Web Development & Frontend": {
		"https://smashingmagazine.com/feed",
		"https://css-tricks.com/feed",
		"https://tympanus.net/codrops/feed",
	},

	"Software Engineering & Architecture": {
		"https://martinfowler.com/feed.atom",
		"https://feeds.dzone.com/home",
		"https://blog.cleancoder.com/atom.xml",
		"https://www.thoughtworks.com/insights.rss",
	},

	"Language-Specific Blogs": {
		"https://blog.rust-lang.org/feed.xml",
		"https://go.dev/blog/feed.atom",
		"https://blog.python.org/feeds/posts/default",
		"https://nodejs.org/en/feed/blog.xml",
		"https://kotlinlang.org/feed.xml",
		"https://www.swift.org/blog/rss.xml",
		"https://blog.jetbrains.com/kotlin/feed/",
		"https://blog.golang.org/feed.atom",
		"https://elixir-lang.org/blog/feed.rss",
		"https://crystal-lang.org/feed.xml",
	},

	"DevOps & Cloud": {
		"https://aws.amazon.com/blogs/aws/feed/",
		"https://cloud.google.com/blog/products/devops-sre/rss",
		"https://kubernetes.io/feed.xml",
		"https://www.docker.com/blog/feed/",
		"https://www.hashicorp.com/blog/feed.xml",
		"https://blog.heroku.com/feed",
		"https://circleci.com/blog/feed.xml",
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
		"https://slack.engineering/feed/",
		"https://shopify.engineering/blog.atom",
		"https://discord.com/blog/rss",
		"https://www.figma.com/blog/feed/",
		"https://blog.mozilla.org/hacks/feed/",
		"https://stackoverflow.blog/engineering/feed/",
		"https://medium.com/feed/square-corner-blog",
	},

	"Reddit Programming": {
		"https://www.reddit.com/r/programming/.rss",
		"https://www.reddit.com/r/webdev/.rss",
		"https://www.reddit.com/r/javascript/.rss",
		"https://www.reddit.com/r/python/.rss",
		"https://www.reddit.com/r/rust/.rss",
		"https://www.reddit.com/r/golang/.rss",
		"https://www.reddit.com/r/node/.rss",
		"https://www.reddit.com/r/reactjs/.rss",
		"https://www.reddit.com/r/cpp/.rss",
		"https://www.reddit.com/r/java/.rss",
		"https://www.reddit.com/r/kubernetes/.rss",
		"https://www.reddit.com/r/devops/.rss",
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
