# Email Reminder System

A web application for creating, managing, and automatically sending scheduled email reminders. Built with Go and SQLite, featuring a Turkish-language interface and secure authentication.

## Features

- Secure password-protected authentication
- Create and manage email reminders
- Flexible scheduling (daily, weekly, monthly, custom intervals)
- Automatic background email sending
- Configurable SMTP settings
- Modern, responsive web interface
- Turkish language support

## Quick Start

### Prerequisites

- Go 1.21 or higher
- SQLite (included via modernc.org/sqlite)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd email-reminder-system
```

2. Install dependencies:
```bash
go mod download
```

3. Create the first user:
```bash
go run cmd/createuser/main.go -i
```

Or use the pre-built executable (if available):
```bash
.\createuser.exe -i
```

Follow the prompts to create your admin user. See [User Creation Guide](docs/USER_CREATION.md) for more details.

4. Start the application:
```bash
go run main.go
```

5. Open your browser and navigate to:
```
http://localhost:8080/login
```

## Configuration

The application can be configured using environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_PATH` | Path to SQLite database file | `data/reminders.db` |
| `ENCRYPTION_KEY` | Base64-encoded 32-byte encryption key | (auto-generated for dev) |
| `SERVER_PORT` | HTTP server port | `8080` |
| `SESSION_DURATION` | Session duration (e.g., "24h", "30m") | `24h` |

Example:
```bash
export DATABASE_PATH="data/reminders.db"
export ENCRYPTION_KEY="your-base64-encoded-32-byte-key"
export SERVER_PORT="8080"
export SESSION_DURATION="24h"
go run main.go
```

## User Management

### Creating Users

Use the `createuser` utility to create new users:

**Interactive mode (recommended):**
```bash
go run cmd/createuser/main.go -i
```

**With username flag:**
```bash
go run cmd/createuser/main.go -username admin
```

For detailed instructions, see the [User Creation Guide](docs/USER_CREATION.md).

### Building the User Creation Utility

Build a standalone executable:
```bash
go build -o createuser.exe cmd/createuser/main.go
```

## Usage

1. **Login**: Access the application at http://localhost:8080/login
2. **Configure SMTP**: Go to Settings and configure your email server settings
3. **Create Reminders**: Click "Hatırlatıcı Ekle" to create a new reminder
4. **Manage Reminders**: View, edit, or delete reminders from the main page

## Project Structure

```
email-reminder-system/
├── cmd/
│   └── createuser/          # User creation utility
├── config/                  # Configuration management
├── database/                # Database layer and migrations
├── handlers/                # HTTP request handlers
├── models/                  # Data models
├── services/                # Business logic services
├── templates/               # HTML templates
├── static/                  # CSS and JavaScript files
├── data/                    # Database storage (created at runtime)
├── docs/                    # Documentation
└── main.go                  # Application entry point
```

## Development

### Running Tests

```bash
go test ./...
```

### Building for Production

```bash
go build -o email-reminder-system.exe
```

## Security

- Passwords are hashed using bcrypt (cost factor 12)
- SMTP passwords are encrypted using AES-256-GCM
- Sessions use secure, random UUIDs
- HTTPS recommended for production deployments
- Secure cookie flags (HttpOnly, Secure, SameSite)

## License

[Your License Here]

## Contributing

[Your Contributing Guidelines Here]

## Support

For issues and questions, please [open an issue](link-to-issues).
