# Migration Guide: TO_EMAIL Removal

## Overview

As of this update, `TO_EMAIL` has been removed from The Paper configuration. All recipients are now managed through the PostgreSQL database.

## What Changed

### Before (Single User)
```bash
# .env
TO_EMAIL=user@example.com
```

The application sent to a single hardcoded email address.

### After (Multi-User Database)
```bash
# .env - TO_EMAIL removed
DB_CONNECT_STRING=postgresql://user:password@localhost:5432/thepaper
```

The application sends to all subscribed users in the `users` table.

## Migration Steps

### Step 1: Set up Database

If you haven't already set up the database:

```bash
# Add to .env
DB_CONNECT_STRING=postgresql://user:password@localhost:5432/thepaper

# Initialize database
cd scripts
go run seed_sources.go  # Seeds all RSS feeds
```

### Step 2: Migrate Your Recipient

Add your existing `TO_EMAIL` recipient to the database:

**Option A: Edit and run the script**

Edit `scripts/add_user.go`:
```go
email := "your-email@example.com"  // Replace with your TO_EMAIL value
name := "Your Name"                // Your name
```

Then run:
```bash
cd scripts
go run add_user.go
```

**Option B: Add directly via SQL**

```sql
INSERT INTO users (email, name, subscribed, unsubscribe_token, created_at, updated_at)
VALUES (
    'your-email@example.com',
    'Your Name',
    true,
    encode(gen_random_bytes(32), 'hex'),  -- PostgreSQL
    NOW(),
    NOW()
);
```

### Step 3: Remove TO_EMAIL from .env

Edit your `.env` file and remove the `TO_EMAIL` line:

```bash
# Remove this line:
# TO_EMAIL=user@example.com
```

### Step 4: Test

```bash
go build -o thepaper
./thepaper
```

You should see:
```
Connecting to database...
✓ Database connection established
Loading configuration...
Found 1 subscribed user(s)
...
✓ Sent successfully to your-email@example.com
```

## Adding More Users

Now that you're on the database system, you can easily add more subscribers:

### Method 1: Via Script
```bash
# Edit scripts/add_user.go with new email/name
cd scripts
go run add_user.go
```

### Method 2: Via SQL
```sql
INSERT INTO users (email, name, subscribed, unsubscribe_token, created_at, updated_at)
VALUES (
    'newuser@example.com',
    'New User',
    true,
    encode(gen_random_bytes(32), 'hex'),
    NOW(),
    NOW()
);
```

### Method 3: Future Web Interface
Once the subscription web interface is built, users can self-subscribe via `/subscribe`.

## Troubleshooting

### "No subscribed users found, exiting"

**Cause:** No users in the database or all users are unsubscribed.

**Solution:**
```sql
-- Check if users exist
SELECT email, subscribed FROM users;

-- If no users, add one
-- (use Method 2 above)

-- If users exist but unsubscribed, resubscribe
UPDATE users SET subscribed = true WHERE email = 'your-email@example.com';
```

### "Failed to connect to database"

**Cause:** `DB_CONNECT_STRING` is missing or incorrect.

**Solution:**
1. Verify `DB_CONNECT_STRING` is in `.env`
2. Check PostgreSQL is running
3. Verify connection string format:
   ```
   postgresql://username:password@host:port/database
   ```

### "TO_EMAIL environment variable is required"

**Cause:** You're running an old version of the code.

**Solution:**
```bash
# Pull latest changes
git pull

# Rebuild
go build -o thepaper
```

## Rollback (If Needed)

If you need to temporarily rollback to the old single-user system:

1. Checkout previous commit:
   ```bash
   git log --oneline  # Find commit before TO_EMAIL removal
   git checkout <commit-hash>
   ```

2. Add `TO_EMAIL` back to `.env`

3. Rebuild:
   ```bash
   go build -o thepaper
   ```

However, this is **not recommended** as you'll lose multi-user support and duplicate prevention.

## Benefits of New System

✅ **Multi-user support**: Send to unlimited subscribers
✅ **Duplicate prevention**: Articles tracked for 30 days
✅ **User management**: Subscribe/unsubscribe workflow
✅ **Analytics ready**: Track opens, engagement (future)
✅ **Scalable**: Add users without code changes
✅ **Compliance ready**: Unsubscribe tokens generated

## Next Steps

1. ✅ Migrate your existing recipient to database
2. ✅ Remove `TO_EMAIL` from `.env`
3. ✅ Test sending
4. 📋 Add additional subscribers as needed
5. 🚀 Enjoy multi-user newsletter management!

## Questions?

- See [DATABASE_SETUP.md](DATABASE_SETUP.md) for full database documentation
- See [scripts/README.md](scripts/README.md) for script usage
- See [README.md](README.md) for general setup

## Summary

| Before | After |
|--------|-------|
| Single recipient in `.env` | Multiple users in database |
| Hardcoded in config | Managed via SQL/scripts |
| No duplicate tracking | 30-day duplicate prevention |
| Manual code changes to add users | Add via scripts or SQL |
| `TO_EMAIL` required | `DB_CONNECT_STRING` required |

**Migration time:** ~5 minutes
**Downtime:** None (if database already set up)