# Design Document: Email Reminder System

## Overview

The Email Reminder System is a Go-based web application that provides authenticated users with the ability to create, manage, and automatically send scheduled email reminders. The system follows a clean architecture pattern with clear separation between HTTP handlers, business logic, data access, and background processing.

The application uses:
- **Go standard library** for HTTP server and routing
- **SQLite** with `database/sql` for data persistence
- **Go templates** for server-side HTML rendering
- **SMTP client** from `net/smtp` for email delivery
- **Session-based authentication** with secure cookie storage
- **Background goroutine** for scheduled email processing

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         Web Browser                          │
└────────────────────────┬────────────────────────────────────┘
                         │ HTTP/HTTPS
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                      HTTP Server (Go)                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ Auth         │  │ Reminder     │  │ Settings     │      │
│  │ Middleware   │  │ Handlers     │  │ Handlers     │      │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘      │
│         │                  │                  │              │
│         ▼                  ▼                  ▼              │
│  ┌────────────────────────────────────────────────────┐    │
│  │           Business Logic Services                   │    │
│  │  - AuthService                                      │    │
│  │  - ReminderService                                  │    │
│  │  - SettingsService                                  │    │
│  └────────────────────┬───────────────────────────────┘    │
│                       │                                      │
│                       ▼                                      │
│  ┌────────────────────────────────────────────────────┐    │
│  │           Data Access Layer (Repository)            │    │
│  └────────────────────┬───────────────────────────────┘    │
└───────────────────────┼──────────────────────────────────────┘
                        │
                        ▼
              ┌──────────────────┐
              │  SQLite Database │
              └──────────────────┘

┌─────────────────────────────────────────────────────────────┐
│              Background Scheduler (Goroutine)                │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  1. Check for due reminders (every minute)           │  │
│  │  2. Load reminder details from database              │  │
│  │  3. Send emails via SMTP                             │  │
│  │  4. Update last-sent timestamp                       │  │
│  │  5. Calculate next send time                         │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### Directory Structure

```
email-reminder-system/
├── main.go                 # Application entry point
├── go.mod                  # Go module definition
├── go.sum                  # Go dependencies
├── config/
│   └── config.go          # Configuration management
├── models/
│   ├── user.go            # User model
│   ├── reminder.go        # Reminder model
│   ├── smtp_settings.go   # SMTP settings model
│   └── session.go         # Session model
├── database/
│   ├── db.go              # Database initialization
│   ├── migrations.go      # Schema migrations
│   └── repository.go      # Data access layer
├── services/
│   ├── auth.go            # Authentication service
│   ├── reminder.go        # Reminder business logic
│   ├── settings.go        # Settings management
│   ├── email.go           # Email sending service
│   └── scheduler.go       # Background scheduler
├── handlers/
│   ├── auth.go            # Login/logout handlers
│   ├── reminder.go        # Reminder CRUD handlers
│   ├── settings.go        # Settings handlers
│   └── middleware.go      # Authentication middleware
├── templates/
│   ├── base.html          # Base template layout
│   ├── login.html         # Login page
│   ├── reminders.html     # Reminder list page
│   ├── reminder_form.html # Reminder create/edit form
│   └── settings.html      # Settings page
├── static/
│   ├── css/
│   │   └── style.css      # Application styles
│   └── js/
│       └── app.js         # Client-side JavaScript
└── data/
    └── reminders.db       # SQLite database file
```

## Components and Interfaces

### 1. Models

#### User Model
```go
type User struct {
    ID           int64
    Username     string
    PasswordHash string
    CreatedAt    time.Time
}
```

#### Reminder Model
```go
type Reminder struct {
    ID           int64
    UserID       int64
    Title        string
    Recipients   []string  // Email addresses
    EmailContent string
    ScheduleType string    // "daily", "weekly", "monthly", "custom"
    IntervalDays int       // For custom schedules
    DayOfWeek    int       // 0-6 for weekly (0=Sunday)
    DayOfMonth   int       // 1-31 for monthly
    TimeOfDay    string    // HH:MM format
    LastSentAt   *time.Time
    NextSendAt   time.Time
    IsActive     bool
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

#### SMTPSettings Model
```go
type SMTPSettings struct {
    ID         int64
    UserID     int64
    Host       string
    Port       int
    Username   string
    Password   string  // Encrypted
    FromEmail  string
    FromName   string
    UseTLS     bool
    UpdatedAt  time.Time
}
```

#### Session Model
```go
type Session struct {
    ID        string
    UserID    int64
    ExpiresAt time.Time
    CreatedAt time.Time
}
```

### 2. Repository Interface

```go
type Repository interface {
    // User operations
    GetUserByUsername(username string) (*User, error)
    CreateUser(user *User) error
    
    // Session operations
    CreateSession(session *Session) error
    GetSession(id string) (*Session, error)
    DeleteSession(id string) error
    CleanExpiredSessions() error
    
    // Reminder operations
    CreateReminder(reminder *Reminder) error
    GetReminder(id int64) (*Reminder, error)
    GetRemindersByUserID(userID int64) ([]*Reminder, error)
    GetDueReminders(now time.Time) ([]*Reminder, error)
    UpdateReminder(reminder *Reminder) error
    DeleteReminder(id int64) error
    
    // SMTP Settings operations
    GetSMTPSettings(userID int64) (*SMTPSettings, error)
    UpsertSMTPSettings(settings *SMTPSettings) error
}
```

### 3. Service Interfaces

#### AuthService
```go
type AuthService interface {
    Login(username, password string) (*Session, error)
    Logout(sessionID string) error
    ValidateSession(sessionID string) (*User, error)
    HashPassword(password string) (string, error)
    VerifyPassword(hash, password string) bool
}
```

#### ReminderService
```go
type ReminderService interface {
    CreateReminder(userID int64, reminder *Reminder) error
    GetReminder(id, userID int64) (*Reminder, error)
    ListReminders(userID int64) ([]*Reminder, error)
    UpdateReminder(reminder *Reminder, userID int64) error
    DeleteReminder(id, userID int64) error
    CalculateNextSendTime(reminder *Reminder) (time.Time, error)
    ValidateReminder(reminder *Reminder) error
}
```

#### EmailService
```go
type EmailService interface {
    SendEmail(to []string, subject, body string, settings *SMTPSettings) error
    ValidateEmailAddress(email string) bool
    ValidateSMTPSettings(settings *SMTPSettings) error
}
```

#### SchedulerService
```go
type SchedulerService interface {
    Start() error
    Stop() error
    ProcessDueReminders() error
}
```

### 4. HTTP Handlers

#### Authentication Handlers
- `GET /login` - Display login page
- `POST /login` - Process login credentials
- `POST /logout` - Invalidate session and logout

#### Reminder Handlers
- `GET /` - Display reminder list (main page)
- `GET /reminders/new` - Display reminder creation form
- `POST /reminders` - Create new reminder
- `GET /reminders/:id/edit` - Display reminder edit form
- `POST /reminders/:id` - Update existing reminder
- `POST /reminders/:id/delete` - Delete reminder

#### Settings Handlers
- `GET /settings` - Display SMTP settings page
- `POST /settings` - Update SMTP settings

## Data Models

### Database Schema

```sql
-- Users table
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Sessions table
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Reminders table
CREATE TABLE reminders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    recipients TEXT NOT NULL,  -- JSON array of email addresses
    email_content TEXT NOT NULL,
    schedule_type TEXT NOT NULL,
    interval_days INTEGER,
    day_of_week INTEGER,
    day_of_month INTEGER,
    time_of_day TEXT NOT NULL,
    last_sent_at DATETIME,
    next_send_at DATETIME NOT NULL,
    is_active BOOLEAN DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- SMTP Settings table
CREATE TABLE smtp_settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    host TEXT NOT NULL,
    port INTEGER NOT NULL,
    username TEXT NOT NULL,
    password TEXT NOT NULL,  -- Encrypted
    from_email TEXT NOT NULL,
    from_name TEXT,
    use_tls BOOLEAN DEFAULT 1,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(user_id)
);

-- Indexes
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX idx_reminders_user_id ON reminders(user_id);
CREATE INDEX idx_reminders_next_send_at ON reminders(next_send_at);
CREATE INDEX idx_smtp_settings_user_id ON smtp_settings(user_id);
```

### Schedule Calculation Logic

The system calculates the next send time based on the schedule type:

**Daily Schedule:**
- Next send = Today at specified time (if not yet passed) OR Tomorrow at specified time

**Weekly Schedule:**
- Next send = Next occurrence of specified day-of-week at specified time

**Monthly Schedule:**
- Next send = Next occurrence of specified day-of-month at specified time
- Handle edge cases (e.g., day 31 in months with fewer days)

**Custom Schedule:**
- Next send = Last sent time + interval_days at specified time
- If never sent, use creation time + interval_days

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*


### Property 1: Unauthenticated Access Redirection
*For any* protected page URL, when accessed without a valid session, the system should redirect to the login page.
**Validates: Requirements 1.1**

### Property 2: Valid Authentication Creates Session
*For any* valid username and password combination, submitting credentials should create a session and grant access to protected pages.
**Validates: Requirements 1.2**

### Property 3: Invalid Authentication Rejection
*For any* invalid credentials (wrong username, wrong password, or malformed input), the login attempt should be rejected with an appropriate error message.
**Validates: Requirements 1.3**

### Property 4: Expired Session Denial
*For any* expired or invalidated session, attempts to access protected pages should be denied and require re-authentication.
**Validates: Requirements 1.4**

### Property 5: Password Hashing Security
*For any* stored user password, the stored value should be a cryptographic hash, not plaintext, and should verify correctly against the original password.
**Validates: Requirements 1.5**

### Property 6: Reminder List Display
*For any* set of reminders belonging to a user, accessing the main page should display all reminders with their titles and schedules.
**Validates: Requirements 2.1**

### Property 7: Reminder Persistence Round-Trip
*For any* valid reminder, creating it and then retrieving it from the database should produce an equivalent reminder with all fields preserved.
**Validates: Requirements 2.3, 7.1**

### Property 8: Reminder Validation Rejection
*For any* reminder missing required fields (title, recipients, content, schedule type, or schedule settings), the system should reject the submission with validation errors indicating which fields are invalid.
**Validates: Requirements 2.4, 3.1, 3.2, 3.3, 3.4, 3.5**

### Property 9: Email Address Validation
*For any* string that is not a valid email address format, the system should reject it as a recipient and indicate the specific invalid addresses.
**Validates: Requirements 3.6**

### Property 10: Reminder Update Preservation
*For any* existing reminder and any valid modifications, updating the reminder should persist all changes to the database while preserving data integrity.
**Validates: Requirements 2.5**

### Property 11: Reminder Deletion Completeness
*For any* reminder, deleting it should remove it from the database such that subsequent queries for that reminder return not found.
**Validates: Requirements 2.6**

### Property 12: SMTP Settings Persistence Round-Trip
*For any* valid SMTP settings, saving them and then retrieving them should produce equivalent settings with all fields preserved (with passwords remaining encrypted).
**Validates: Requirements 4.3, 7.2**

### Property 13: SMTP Settings Validation
*For any* SMTP settings missing required fields (host, port, username, password, from_email) or with invalid values (invalid port numbers, malformed emails), the system should reject the submission with appropriate validation errors.
**Validates: Requirements 4.2, 4.5**

### Property 14: SMTP Password Encryption
*For any* SMTP settings containing a password, the stored password should be encrypted and not plaintext, and should decrypt correctly for use.
**Validates: Requirements 4.4**

### Property 15: Schedule Type Validation
*For any* schedule configuration, the system should validate that required fields for that schedule type are present: daily requires time-of-day, weekly requires day-of-week and time-of-day, monthly requires day-of-month and time-of-day, and custom requires interval-days and time-of-day.
**Validates: Requirements 6.1, 6.2, 6.3, 6.4**

### Property 16: Time Format Validation
*For any* time-of-day value that is not in valid 24-hour format (HH:MM where HH is 00-23 and MM is 00-59), the system should reject it with a validation error.
**Validates: Requirements 6.5**

### Property 17: Next Send Time Calculation
*For any* reminder with a valid schedule configuration and current time, calculating the next send time should produce a future timestamp that correctly reflects the schedule rules (daily = next occurrence at specified time, weekly = next specified day-of-week at time, monthly = next specified day-of-month at time, custom = current time + interval days at time).
**Validates: Requirements 6.6**

### Property 18: Due Reminder Detection
*For any* set of reminders, the scheduler should identify all reminders whose next_send_at time is less than or equal to the current time as due for sending.
**Validates: Requirements 5.1**

### Property 19: Email Sending to All Recipients
*For any* due reminder with multiple recipients, the system should send an email to each recipient in the list using the configured SMTP settings.
**Validates: Requirements 5.2, 5.3**

### Property 20: Last-Sent Timestamp Update
*For any* reminder that successfully sends emails, the system should update the last_sent_at timestamp to the current time.
**Validates: Requirements 5.4**

### Property 21: Next Send Time Recalculation
*For any* recurring reminder that successfully sends, the system should calculate and update the next_send_at time based on the schedule configuration.
**Validates: Requirements 5.6**

### Property 22: Database Schema Idempotency
*For any* number of application starts, initializing the database schema should result in the same schema structure without errors, whether the database is new or already exists.
**Validates: Requirements 7.4**

### Property 23: Data Persistence Across Restarts
*For any* reminder or SMTP settings created before an application restart, the data should be retrievable after the restart with all fields intact.
**Validates: Requirements 7.5**

### Property 24: Referential Integrity Enforcement
*For any* attempt to create or modify data that would violate foreign key constraints (e.g., reminder with non-existent user_id), the database should reject the operation.
**Validates: Requirements 7.6**

### Property 25: Reminder Display Completeness
*For any* reminder in the list view, the rendered output should contain the reminder's title, schedule type, and next send time.
**Validates: Requirements 8.3**

### Property 26: Validation Feedback Display
*For any* form submission with validation errors, the system should display error messages indicating which fields are invalid and why.
**Validates: Requirements 8.4, 8.6**

### Property 27: Comprehensive Error Logging
*For any* error that occurs (email sending failure, database error, authentication failure), the system should log the error with appropriate level (ERROR for failures, WARN for retryable issues, INFO for successful operations), timestamp, and relevant context (user ID, reminder ID, error details).
**Validates: Requirements 9.1, 9.2, 9.4, 9.5**

## Error Handling

### Authentication Errors
- **Invalid credentials**: Return 401 Unauthorized with error message
- **Expired session**: Redirect to login page with session expired message
- **Missing session**: Redirect to login page

### Validation Errors
- **Missing required fields**: Return 400 Bad Request with field-specific error messages
- **Invalid email format**: Return 400 Bad Request with list of invalid emails
- **Invalid time format**: Return 400 Bad Request with format requirements
- **Invalid schedule configuration**: Return 400 Bad Request with missing field details

### Database Errors
- **Connection failure**: Log error, return 500 Internal Server Error
- **Constraint violation**: Return 400 Bad Request with constraint details
- **Query timeout**: Log error, retry once, return 500 if retry fails

### Email Sending Errors
- **SMTP connection failure**: Log error with connection details, mark reminder for retry
- **Authentication failure**: Log error, notify user to check SMTP settings
- **Recipient rejection**: Log error with rejected addresses, continue sending to valid addresses
- **Timeout**: Log error, implement exponential backoff retry (3 attempts)

### Retry Policy
- **Email sending failures**: Retry up to 3 times with exponential backoff (1 min, 5 min, 15 min)
- **Database connection failures**: Retry up to 3 times with 1-second delay
- **Failed reminders**: Mark as failed after max retries, log for manual review

## Testing Strategy

### Dual Testing Approach

The system will use both unit tests and property-based tests for comprehensive coverage:

**Unit Tests** focus on:
- Specific examples of correct behavior
- Edge cases (e.g., day 31 in February for monthly schedules)
- Error conditions (e.g., SMTP connection failures)
- Integration points between components

**Property-Based Tests** focus on:
- Universal properties that hold for all inputs
- Comprehensive input coverage through randomization
- Validating correctness properties from this design document

### Property-Based Testing Configuration

**Library**: Use `gopter` (Go property testing library) for property-based tests

**Configuration**:
- Minimum 100 iterations per property test
- Each test must reference its design document property
- Tag format: `// Feature: email-reminder-system, Property N: [property text]`

**Example Property Test Structure**:
```go
// Feature: email-reminder-system, Property 7: Reminder Persistence Round-Trip
func TestReminderPersistenceRoundTrip(t *testing.T) {
    properties := gopter.NewProperties(nil)
    properties.Property("reminder round-trip preserves all fields", 
        prop.ForAll(
            func(reminder *Reminder) bool {
                // Create reminder in database
                err := repo.CreateReminder(reminder)
                if err != nil {
                    return false
                }
                
                // Retrieve reminder from database
                retrieved, err := repo.GetReminder(reminder.ID)
                if err != nil {
                    return false
                }
                
                // Verify all fields match
                return remindersEqual(reminder, retrieved)
            },
            genValidReminder(), // Generator for valid reminders
        ))
    
    properties.TestingRun(t, gopter.ConsoleReporter(false))
}
```

### Test Coverage Requirements

**Unit Tests**:
- Authentication flow (login, logout, session validation)
- Reminder CRUD operations with specific examples
- Schedule calculation edge cases (month boundaries, leap years)
- Email validation with various formats
- SMTP connection error handling
- Database migration and initialization

**Property-Based Tests** (one test per property):
- Property 1-27 as defined in Correctness Properties section
- Each property test runs minimum 100 iterations
- Generators for valid and invalid inputs
- Custom generators for complex types (Reminder, SMTPSettings, Schedule)

### Integration Tests

- End-to-end reminder creation and email sending
- Authentication flow with database persistence
- Scheduler processing due reminders
- SMTP settings update and email sending with new settings

### Test Data Generators

**For Property-Based Tests**:
- `genValidReminder()`: Generates reminders with all required fields
- `genInvalidReminder()`: Generates reminders missing required fields
- `genValidEmail()`: Generates valid email addresses
- `genInvalidEmail()`: Generates invalid email formats
- `genSchedule()`: Generates valid schedule configurations
- `genSMTPSettings()`: Generates valid SMTP configurations
- `genTimeOfDay()`: Generates valid HH:MM time strings

### Testing Best Practices

1. **Isolation**: Each test should be independent and not rely on other tests
2. **Cleanup**: Tests should clean up database state after execution
3. **Mocking**: Use mocks for SMTP client in unit tests to avoid sending real emails
4. **Time Handling**: Use dependency injection for time.Now() to enable deterministic testing
5. **Database**: Use in-memory SQLite for faster test execution
6. **Parallel Execution**: Unit tests should be safe for parallel execution
7. **Property Test Balance**: Focus property tests on core business logic, use unit tests for edge cases

## Security Considerations

### Authentication
- Passwords hashed using bcrypt with cost factor 12
- Sessions stored with secure, random IDs (UUID v4)
- Session expiration after 24 hours of inactivity
- CSRF protection for all state-changing operations

### Data Protection
- SMTP passwords encrypted using AES-256-GCM
- Encryption key stored in environment variable, never in code
- Database file permissions restricted to application user only
- No sensitive data in logs (passwords, session IDs masked)

### Input Validation
- All user inputs validated and sanitized
- Email addresses validated against RFC 5322 format
- SQL injection prevention through parameterized queries
- XSS prevention through template auto-escaping

### Network Security
- HTTPS enforced in production (TLS 1.2+)
- SMTP connections use TLS when available
- Secure cookie flags (HttpOnly, Secure, SameSite)
- Rate limiting on login attempts (5 attempts per 15 minutes)

## Performance Considerations

### Database Optimization
- Indexes on frequently queried columns (user_id, next_send_at, expires_at)
- Connection pooling with max 10 connections
- Prepared statements for repeated queries
- Batch operations for bulk updates

### Scheduler Efficiency
- Check for due reminders every 60 seconds
- Process reminders in batches of 10
- Concurrent email sending with worker pool (5 workers)
- Timeout of 30 seconds per email send operation

### Caching Strategy
- Cache SMTP settings in memory (invalidate on update)
- Cache user session data in memory with TTL
- No caching of reminder list (always fresh data)

### Resource Limits
- Maximum 100 recipients per reminder
- Maximum 10KB email content size
- Maximum 50 active reminders per user
- Session cleanup runs every hour to remove expired sessions
