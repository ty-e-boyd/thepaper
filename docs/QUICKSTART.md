# Quick Start Guide - Newsletter Subscription System

Get the subscribe/unsubscribe functionality up and running in 5 minutes.

## Prerequisites

- PostgreSQL running
- Go 1.22+ installed
- Both projects cloned

## 1. Database Setup (30 seconds)

```bash
# Create database if it doesn't exist
createdb thepaper

# Or with psql
psql -c "CREATE DATABASE thepaper;"
```

## 2. Configure Environment Variables (1 minute)

**thepaper/.env:**
```bash
DB_CONNECT_STRING=postgresql://user:password@localhost:5432/thepaper?sslmode=disable
PORTFOLIO_URL=http://localhost:4040
GEMINI_API_KEY=your_gemini_key
SENDGRID_API_KEY=your_sendgrid_key
FROM_EMAIL=newsletter@example.com
```

**tyler-portfolio/.env:**
```bash
DB_CONNECT_STRING=postgresql://user:password@localhost:5432/thepaper?sslmode=disable
```

⚠️ **Important:** Both must use the SAME database connection string!

## 3. Start the Portfolio (30 seconds)

```bash
cd go_projects/tyler-portfolio
go run main.go
```

Should see:
```
✓ Database connection established
⇨ http server started on [::]:4040
```

## 4. Test Subscribe (1 minute)

Open browser: `http://localhost:4040/subscribe`

Enter an email and click Subscribe.

**Verify in database:**
```bash
psql -d thepaper -c "SELECT email, subscribed, created_at FROM users;"
```

Should see your email with `subscribed = true` and a timestamp.

## 5. Fix Existing Users (if needed)

If you have users with NULL timestamps:

```bash
cd thepaper
psql -d thepaper -f scripts/fix_timestamps.sql
```

## 6. Test Newsletter Send (1 minute)

```bash
cd thepaper

# Dry run (no emails sent)
go run main.go --dry-run

# Should show your subscribed user in the list
```

## 7. Send Test Email (optional)

```bash
cd thepaper
go run main.go
```

Check your inbox for the newsletter with an unsubscribe link.

## 8. Test Unsubscribe (30 seconds)

Click the unsubscribe link in the email, or manually:

```bash
# Get token from database
TOKEN=$(psql -d thepaper -t -c "SELECT unsubscribe_token FROM users LIMIT 1;" | tr -d ' ')

# Visit unsubscribe URL
curl "http://localhost:4040/unsubscribe?token=$TOKEN"
```

**Verify in database:**
```bash
psql -d thepaper -c "SELECT email, subscribed, updated_at FROM users;"
```

Should show `subscribed = false` and `updated_at` timestamp updated.

## Common Commands

**Check all users:**
```bash
psql -d thepaper -c "SELECT id, email, subscribed, created_at FROM users ORDER BY created_at DESC;"
```

**Count subscribers:**
```bash
psql -d thepaper -c "SELECT COUNT(*) FROM users WHERE subscribed = true;"
```

**Resubscribe a user:**
```bash
psql -d thepaper -c "UPDATE users SET subscribed = true WHERE email = 'user@example.com';"
```

**Get unsubscribe token:**
```bash
psql -d thepaper -c "SELECT unsubscribe_token FROM users WHERE email = 'user@example.com';"
```

## Troubleshooting

### Portfolio won't start
- Check PostgreSQL is running: `psql -d thepaper -c "SELECT 1;"`
- Verify `.env` file exists and has `DB_CONNECT_STRING`

### User has NULL timestamps
```bash
cd thepaper
psql -d thepaper -f scripts/fix_timestamps.sql
```

### Can't subscribe (email already exists)
- User might be unsubscribed - just resubscribe from the form
- Or manually: `UPDATE users SET subscribed = true WHERE email = 'user@example.com';`

### Unsubscribe link doesn't work
- Check `PORTFOLIO_URL` is set in `thepaper/.env`
- Verify token in database matches token in URL
- Ensure both apps use same database

## Production Setup

1. **Set production URLs:**
```bash
# thepaper/.env
PORTFOLIO_URL=https://tyboyd.dev
```

2. **Use SSL for database:**
```bash
DB_CONNECT_STRING=postgresql://user:password@prod-db:5432/thepaper?sslmode=require
```

3. **Deploy portfolio** behind reverse proxy (nginx/caddy) with SSL

4. **Schedule thepaper** with cron:
```bash
0 8 * * * cd /path/to/thepaper && ./thepaper >> /var/log/thepaper.log 2>&1
```

## That's It! 🎉

You now have a fully functional newsletter subscription system with:
- ✅ Subscribe page
- ✅ Email collection
- ✅ Unsubscribe links in emails
- ✅ Resubscribe capability
- ✅ Secure token generation

## Next Steps

- Read `SUBSCRIPTION_SETUP.md` for detailed documentation
- Review `TROUBLESHOOTING.md` for common issues
- Check `README.md` for API reference
- Customize the subscribe page design in `public/subscribe.html`

## Support

Having issues? Check:
1. `TROUBLESHOOTING.md` in tyler-portfolio
2. `SUBSCRIPTION_SETUP.md` for detailed setup
3. Application logs for error messages
4. Database connectivity with `psql`
