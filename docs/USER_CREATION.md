# User Creation Guide

This guide explains how to create users for the Email Reminder System.

## Overview

The Email Reminder System requires users to be created before they can log in. This is done using the `createuser` command-line utility.

## Creating the First User

### Method 1: Interactive Mode (Recommended)

The interactive mode prompts you for username and password securely:

```bash
go run cmd/createuser/main.go -i
```

You will be prompted to:
1. Enter a username
2. Enter a password (hidden input)
3. Confirm the password

Example:
```
Enter username: admin
Enter password: 
Confirm password: 
✓ User 'admin' created successfully (ID: 1)

You can now log in to the Email Reminder System with these credentials.
```

### Method 2: Command-Line with Username Flag

Provide the username via flag and be prompted for password:

```bash
go run cmd/createuser/main.go -username admin
```

You will be prompted to enter the password securely.

### Method 3: Full Command-Line (Not Recommended)

**Warning:** This method is insecure as the password will be visible in command history.

```bash
go run cmd/createuser/main.go -username admin -password mypassword
```

## Building the Utility

You can build a standalone executable for easier use:

```bash
go build -o createuser.exe cmd/createuser/main.go
```

Then run it:

```bash
./createuser.exe -i
```

## Custom Database Path

If you're using a custom database location, specify it with the `-db` flag:

```bash
go run cmd/createuser/main.go -i -db /path/to/custom/database.db
```

## Creating Additional Users

To create additional users, simply run the utility again with a different username:

```bash
go run cmd/createuser/main.go -i
```

The utility will prevent you from creating duplicate usernames.

## Command-Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `-i` | Interactive mode (prompts for credentials) | false |
| `-username` | Username for the new user | (required if not interactive) |
| `-password` | Password for the new user (not recommended) | (prompted if not provided) |
| `-db` | Path to database file | From environment or config |

## Security Notes

1. **Always use interactive mode** (`-i`) or omit the `-password` flag to avoid exposing passwords in command history
2. Passwords are hashed using bcrypt with cost factor 12 before storage
3. The utility checks for existing usernames to prevent duplicates
4. Password confirmation is required in interactive mode to prevent typos

## Troubleshooting

### Error: "Failed to initialize database"
- Ensure the database path is correct and the directory exists
- Check file permissions on the data directory

### Error: "User already exists"
- The username is already taken
- Choose a different username or use the existing credentials

### Error: "Passwords do not match"
- In interactive mode, the password and confirmation didn't match
- Try again and ensure you type the same password twice

## Integration with Application

Once users are created, they can log in to the web application at:
```
http://localhost:8080/login
```

Use the username and password you created with this utility.

## Programmatic User Creation

If you need to create users programmatically (e.g., in tests or initialization scripts), you can use the repository and auth service directly:

```go
import (
    "email-reminder-system/database"
    "email-reminder-system/models"
    "email-reminder-system/services"
)

// Initialize database and services
db, _ := database.InitDB("data/reminders.db")
repo := database.NewSQLiteRepository(db)
authService := services.NewAuthService(repo, 24*time.Hour)

// Hash password
passwordHash, _ := authService.HashPassword("mypassword")

// Create user
user := &models.User{
    Username:     "admin",
    PasswordHash: passwordHash,
}
repo.CreateUser(user)
```

## Next Steps

After creating your first user:
1. Start the Email Reminder System: `go run main.go`
2. Navigate to http://localhost:8080/login
3. Log in with your credentials
4. Configure SMTP settings in the Settings page
5. Create your first reminder
