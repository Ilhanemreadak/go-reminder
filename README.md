# Email Reminder System

A web application for creating, managing, and automatically sending scheduled email reminders. Built with Go and SQLite, featuring a Turkish-language interface and secure authentication.

## Features

- Secure, bcrypt-based password authentication
- Create and manage email reminders (CRUD)
- Flexible scheduling (daily, weekly, monthly, custom intervals)
- Automatic background email sending (checks every 60 seconds)
- Exponential backoff retry mechanism for failed deliveries
- Configurable SMTP settings with connection testing
- Email delivery history with filtering, pagination, and detail view
- Reminder active/inactive toggle
- Structured logging to console and file, with sensitive data masking
- Modern, responsive web interface
- Turkish language support

## Quick Start

### Prerequisites

- Go 1.21 or higher
- SQLite (included via modernc.org/sqlite, no external installation required)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd time-sheet-reminder
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

Follow the prompts to create your admin user. See the [User Creation Guide](docs/USER_CREATION.md) for more details.

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
| `SESSION_DURATION` | Session duration (in hours) | `24` (24 hours) |

Example:
```bash
export DATABASE_PATH="data/reminders.db"
export ENCRYPTION_KEY="your-base64-encoded-32-byte-key"
export SERVER_PORT="8080"
export SESSION_DURATION="24"
go run main.go
```

> **Warning:** If `ENCRYPTION_KEY` is not set, the application uses a default key for development. Always set a strong key in production environments.

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

## CLI Tools

The project includes command-line tools for various management tasks:

| Tool | Description | Usage |
|------|-------------|-------|
| `createuser` | Create a new user | `go run cmd/createuser/main.go -i` |
| `activatereminder` | Activate a reminder | `go run cmd/activatereminder/main.go` |
| `debugreminders` | Debug reminders | `go run cmd/debugreminders/main.go` |
| `verifyuser` | Verify a user | `go run cmd/verifyuser/main.go` |

## Usage

1. **Login**: Access the application at http://localhost:8080/login
2. **Configure SMTP**: Go to Settings and configure your email server settings
3. **Test SMTP**: Use the "Test" button on the Settings page to verify your SMTP connection
4. **Create Reminders**: Click "Hatırlatıcı Ekle" to create a new reminder
5. **Manage Reminders**: View, edit, toggle active/inactive, or delete reminders from the main page
6. **Email History**: View sent emails filtered by status, date range, and reminder

## API Routes

### Public Routes

| Method | Path | Description |
|--------|------|-------------|
| GET/POST | `/login` | Login page and authentication |
| GET | `/logout` | Logout |
| GET | `/health` | Health check endpoint |

### Protected (Authenticated) Routes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | Reminder list |
| GET | `/reminders/new` | New reminder form |
| POST | `/reminders` | Create reminder |
| GET | `/reminders/:id/edit` | Edit reminder form |
| POST | `/reminders/:id` | Update reminder |
| POST | `/reminders/:id/delete` | Delete reminder |
| POST | `/reminders/:id/toggle` | Toggle reminder active/inactive |
| GET/POST | `/settings` | SMTP settings page and update |
| POST | `/settings/test` | SMTP connection test |
| GET | `/email-history` | Email history (filtering & pagination) |
| GET | `/email-history/:id` | Email detail (JSON) |

## Project Structure

```
time-sheet-reminder/
├── cmd/
│   ├── activatereminder/    # Reminder activation tool
│   ├── createuser/          # User creation tool
│   ├── debugreminders/      # Reminder debugging tool
│   └── verifyuser/          # User verification tool
├── config/
│   └── config.go            # Configuration management via environment variables
├── database/
│   ├── db.go                # Database connection management
│   ├── migrations.go        # Database schema migrations
│   └── repository.go        # Data access layer (Repository pattern)
├── handlers/
│   ├── auth.go              # Authentication handlers
│   ├── email_history.go     # Email history handlers
│   ├── middleware.go        # Authentication and recovery middleware
│   ├── reminder.go          # Reminder CRUD handlers
│   ├── reminder_test.go     # Reminder handler tests
│   └── settings.go          # SMTP settings handlers
├── models/
│   ├── email_log.go         # Email log model
│   ├── reminder.go          # Reminder model
│   ├── session.go           # Session model
│   ├── smtp_settings.go     # SMTP settings model
│   └── user.go              # User model
├── services/
│   ├── auth.go              # Authentication service
│   ├── auth_test.go         # Authentication tests
│   ├── email.go             # Email sending service (SMTP)
│   ├── email_history_service.go # Email history service
│   ├── encryption.go        # AES-256-GCM encryption service
│   ├── encryption_test.go   # Encryption tests
│   ├── logger.go            # Structured logging service
│   ├── logger_test.go       # Logger tests
│   ├── reminder.go          # Reminder business logic service
│   ├── reminder_test.go     # Reminder service tests
│   ├── scheduler.go         # Background scheduler service
│   ├── settings.go          # Settings service
│   └── settings_test.go     # Settings service tests
├── templates/
│   ├── base.html            # Base page template (layout)
│   ├── email_history.html   # Email history page
│   ├── login.html           # Login page
│   ├── reminder_form.html   # Reminder create/edit form
│   ├── reminders.html       # Reminder list page
│   └── settings.html        # SMTP settings page
├── static/
│   ├── css/                 # Stylesheets
│   └── js/                  # JavaScript files
├── data/                    # Database storage (created at runtime)
├── logs/                    # Application log files (created at runtime)
├── docs/                    # Documentation
│   └── USER_CREATION.md     # User creation guide
├── .gitignore               # Git ignore rules
├── go.mod                   # Go module definition
├── go.sum                   # Go dependency checksums
└── main.go                  # Application entry point
```

## Architecture

The application follows layered architecture principles:

- **`main.go`** — Application entry point, dependency injection (DI), route definitions, and graceful shutdown
- **`config/`** — Configuration loading from environment variables
- **`handlers/`** — HTTP request handling layer (Controller), middleware (auth & panic recovery)
- **`services/`** — Business logic layer: authentication, email sending, scheduling, encryption, logging
- **`models/`** — Data models (Reminder, User, Session, SMTPSettings, EmailLog)
- **`database/`** — Data access layer: SQLite connection, migrations, repository pattern

### Scheduler

The background scheduler runs every **60 seconds**, checking for due reminders and sending emails. Failed deliveries are retried with increasing delays of **1min → 5min → 15min**. If all attempts fail, the reminder is deactivated.

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

- Passwords are hashed using **bcrypt** (cost factor 12)
- SMTP passwords are encrypted using **AES-256-GCM**
- Sessions use secure, random **UUIDs**
- Secure cookie flags (**HttpOnly**, **Secure**, **SameSite**)
- Structured logging with automatic masking of sensitive data (passwords, sessions, keys)
- Panic recovery middleware catches unexpected errors
- HTTPS recommended for production deployments
- XSS protection: HTML escaping applied to email log details

## Support

For issues and questions, please [open an issue](link-to-issues).
