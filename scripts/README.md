# Scripts

This directory contains utility scripts for managing The Paper database.

## Available Scripts

### 1. seed_sources.go
Seeds all RSS feed sources from `feeds/sources.go` into the database.

**Usage:**
```bash
cd scripts
go run seed_sources.go
```

**What it does:**
- Connects to the database
- Runs migrations to ensure tables exist
- Iterates through all categories and feeds in `feeds.FeedSources`
- Creates a database record for each feed URL
- Skips feeds that already exist (prevents duplicates)
- Shows a summary of sources added/skipped

**When to run:**
- First time setting up the database
- After adding new feeds to `feeds/sources.go`

### 2. add_user.go
Adds a user to the database (or shows existing user info if already exists).

**Usage:**
```bash
cd scripts
go run add_user.go
```

**What it does:**
- Connects to the database
- Runs migrations to ensure tables exist
- Creates a user with email `tyler@tylerevan.dev`
- Generates a unique unsubscribe token
- Shows user information including ID and token

**When to run:**
- First time setting up the database
- To add additional users (modify the script with different email/name)

**To add more users:**
Edit the script and change these lines:
```go
email := "newuser@example.com"
name := "User Name"
```

## Environment Requirements

Both scripts require the `DB_CONNECT_STRING` environment variable to be set in your `.env` file:

```bash
DB_CONNECT_STRING=postgresql://user:password@localhost:5432/thepaper
```

## First-Time Setup

To set up the database from scratch:

1. Ensure PostgreSQL is running
2. Create a database for The Paper
3. Set `DB_CONNECT_STRING` in `.env`
4. Run the scripts in order:
   ```bash
   cd scripts
   go run seed_sources.go
   go run add_user.go
   ```

## Notes

- Both scripts use `godotenv.Load("../.env")` to load environment variables from the parent directory
- Migrations run automatically when scripts connect to the database
- Scripts are idempotent - safe to run multiple times
- GORM handles connection pooling and SQL injection protection