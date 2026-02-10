# Implementation Plan: Email Reminder System

## Overview

This implementation plan breaks down the Email Reminder System into discrete, incremental coding tasks. Each task builds on previous work, starting with foundational components (database, models) and progressing through services, handlers, UI, and finally the background scheduler. The plan includes property-based tests and unit tests as sub-tasks to validate correctness early and often.

## Tasks

- [x] 1. Initialize project structure and dependencies
  - Create Go module with `go mod init email-reminder-system`
  - Add dependencies: `database/sql`, `modernc.org/sqlite`, `golang.org/x/crypto/bcrypt`, `github.com/google/uuid`, `github.com/leanovate/gopter`
  - Create directory structure: `config/`, `models/`, `database/`, `services/`, `handlers/`, `templates/`, `static/css/`, `static/js/`, `data/`
  - Create main.go entry point with basic HTTP server setup
  - _Requirements: 10.1, 10.2, 10.3, 10.5_

- [x] 2. Implement database layer and models
  - [x] 2.1 Create database schema and migrations
    - Write SQL schema for users, sessions, reminders, smtp_settings tables
    - Implement database initialization function that creates tables if they don't exist
    - Add indexes for performance (user_id, next_send_at, expires_at)
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.6_
  
  - [ ]* 2.2 Write property test for database schema idempotency
    - **Property 22: Database Schema Idempotency**
    - **Validates: Requirements 7.4**
  
  - [x] 2.3 Create model structs and repository interface
    - Define User, Reminder, SMTPSettings, Session structs in models package
    - Implement Repository interface with all CRUD operations
    - Implement SQLite repository with parameterized queries
    - _Requirements: 7.1, 7.2, 7.3, 7.6_
  
  - [ ]* 2.4 Write property test for reminder persistence round-trip
    - **Property 7: Reminder Persistence Round-Trip**
    - **Validates: Requirements 2.3, 7.1**
  
  - [ ]* 2.5 Write property test for SMTP settings persistence round-trip
    - **Property 12: SMTP Settings Persistence Round-Trip**
    - **Validates: Requirements 4.3, 7.2**
  
  - [ ]* 2.6 Write property test for referential integrity enforcement
    - **Property 24: Referential Integrity Enforcement**
    - **Validates: Requirements 7.6**

- [x] 3. Implement authentication service
  - [x] 3.1 Create authentication service with password hashing
    - Implement AuthService with bcrypt password hashing (cost 12)
    - Implement session creation with UUID generation
    - Implement session validation and expiration checking
    - _Requirements: 1.2, 1.4, 1.5_
  
  - [ ]* 3.2 Write property test for password hashing security
    - **Property 5: Password Hashing Security**
    - **Validates: Requirements 1.5**
  
  - [ ]* 3.3 Write property test for valid authentication creates session
    - **Property 2: Valid Authentication Creates Session**
    - **Validates: Requirements 1.2**
  
  - [ ]* 3.4 Write property test for invalid authentication rejection
    - **Property 3: Invalid Authentication Rejection**
    - **Validates: Requirements 1.3**
  
  - [ ]* 3.5 Write property test for expired session denial
    - **Property 4: Expired Session Denial**
    - **Validates: Requirements 1.4**
  
  - [ ]* 3.6 Write unit tests for authentication edge cases
    - Test session expiration boundary conditions
    - Test concurrent session creation
    - _Requirements: 1.2, 1.4_

- [x] 4. Implement authentication middleware and handlers
  - [x] 4.1 Create authentication middleware
    - Implement middleware that checks for valid session cookie
    - Redirect to login page if session is invalid or missing
    - Store user information in request context for authenticated requests
    - _Requirements: 1.1, 1.4_
  
  - [ ]* 4.2 Write property test for unauthenticated access redirection
    - **Property 1: Unauthenticated Access Redirection**
    - **Validates: Requirements 1.1**
  
  - [x] 4.3 Create login and logout handlers
    - Implement GET /login handler to display login page
    - Implement POST /login handler to process credentials
    - Implement POST /logout handler to invalidate session
    - Set secure session cookies (HttpOnly, Secure, SameSite)
    - _Requirements: 1.2, 1.3, 1.4_
  
  - [ ]* 4.4 Write unit tests for login/logout flow
    - Test successful login flow
    - Test logout clears session
    - Test cookie security flags
    - _Requirements: 1.2, 1.4_

- [x] 5. Checkpoint - Ensure authentication tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 6. Implement reminder service with validation
  - [x] 6.1 Create reminder validation logic
    - Implement validation for required fields (title, recipients, content, schedule)
    - Implement email address validation using regex (RFC 5322)
    - Implement schedule-specific validation (daily, weekly, monthly, custom)
    - Implement time format validation (HH:MM in 24-hour format)
    - _Requirements: 2.4, 3.1, 3.2, 3.3, 3.4, 3.5, 3.6, 6.1, 6.2, 6.3, 6.4, 6.5_
  
  - [ ]* 6.2 Write property test for reminder validation rejection
    - **Property 8: Reminder Validation Rejection**
    - **Validates: Requirements 2.4, 3.1, 3.2, 3.3, 3.4, 3.5**
  
  - [ ]* 6.3 Write property test for email address validation
    - **Property 9: Email Address Validation**
    - **Validates: Requirements 3.6**
  
  - [ ]* 6.4 Write property test for schedule type validation
    - **Property 15: Schedule Type Validation**
    - **Validates: Requirements 6.1, 6.2, 6.3, 6.4**
  
  - [ ]* 6.5 Write property test for time format validation
    - **Property 16: Time Format Validation**
    - **Validates: Requirements 6.5**
  
  - [x] 6.6 Implement schedule calculation logic
    - Implement CalculateNextSendTime function for all schedule types
    - Handle edge cases (month boundaries, leap years, day 31 in shorter months)
    - Calculate initial next_send_at on reminder creation
    - _Requirements: 6.6, 5.6_
  
  - [ ]* 6.7 Write property test for next send time calculation
    - **Property 17: Next Send Time Calculation**
    - **Validates: Requirements 6.6**
  
  - [ ]* 6.8 Write unit tests for schedule calculation edge cases
    - Test monthly schedule with day 31 in February
    - Test weekly schedule across year boundaries
    - Test custom schedule with various intervals
    - _Requirements: 6.6_
  
  - [x] 6.9 Implement ReminderService with CRUD operations
    - Implement CreateReminder with validation and next_send_at calculation
    - Implement GetReminder, ListReminders, UpdateReminder, DeleteReminder
    - Ensure user authorization (users can only access their own reminders)
    - _Requirements: 2.1, 2.3, 2.5, 2.6_
  
  - [ ]* 6.10 Write property test for reminder update preservation
    - **Property 10: Reminder Update Preservation**
    - **Validates: Requirements 2.5**
  
  - [ ]* 6.11 Write property test for reminder deletion completeness
    - **Property 11: Reminder Deletion Completeness**
    - **Validates: Requirements 2.6**

- [x] 7. Implement SMTP settings service
  - [x] 7.1 Create encryption utilities for SMTP passwords
    - Implement AES-256-GCM encryption/decryption functions
    - Load encryption key from environment variable
    - _Requirements: 4.4_
  
  - [ ]* 7.2 Write property test for SMTP password encryption
    - **Property 14: SMTP Password Encryption**
    - **Validates: Requirements 4.4**
  
  - [x] 7.3 Implement SMTP settings validation
    - Validate required fields (host, port, username, password, from_email)
    - Validate port number range (1-65535)
    - Validate from_email format
    - _Requirements: 4.2, 4.5_
  
  - [ ]* 7.4 Write property test for SMTP settings validation
    - **Property 13: SMTP Settings Validation**
    - **Validates: Requirements 4.2, 4.5**
  
  - [x] 7.5 Implement SettingsService with upsert operation
    - Implement GetSMTPSettings and UpsertSMTPSettings
    - Encrypt password before storage, decrypt on retrieval
    - Ensure user authorization
    - _Requirements: 4.1, 4.3, 4.4_

- [x] 8. Implement email sending service
  - [x] 8.1 Create email service with SMTP client
    - Implement SendEmail function using net/smtp
    - Support TLS connections
    - Handle authentication with SMTP server
    - Format email with proper headers (From, To, Subject, Body)
    - _Requirements: 5.2, 5.3_
  
  - [ ]* 8.2 Write unit tests for email service with mock SMTP
    - Test email formatting
    - Test TLS connection setup
    - Test authentication flow
    - Test error handling for connection failures
    - _Requirements: 5.2, 5.3, 9.3_

- [x] 9. Checkpoint - Ensure service layer tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 10. Implement reminder handlers and UI
  - [x] 10.1 Create reminder list handler and template
    - Implement GET / handler to display reminders
    - Create reminders.html template with Turkish labels ("Hatırlatıcılar")
    - Display title, schedule type, next send time for each reminder
    - Add "Hatırlatıcı Ekle" button
    - _Requirements: 2.1, 8.1, 8.3_
  
  - [ ]* 10.2 Write property test for reminder list display
    - **Property 6: Reminder List Display**
    - **Validates: Requirements 2.1**
  
  - [ ]* 10.3 Write property test for reminder display completeness
    - **Property 25: Reminder Display Completeness**
    - **Validates: Requirements 8.3**
  
  - [x] 10.4 Create reminder form handler and template
    - Implement GET /reminders/new handler to display creation form
    - Implement GET /reminders/:id/edit handler to display edit form
    - Create reminder_form.html template with all fields
    - Add client-side validation hints
    - _Requirements: 2.2, 3.1, 3.2, 3.3, 3.4, 3.5_
  
  - [x] 10.5 Create reminder create/update/delete handlers
    - Implement POST /reminders handler to create reminder
    - Implement POST /reminders/:id handler to update reminder
    - Implement POST /reminders/:id/delete handler to delete reminder
    - Display validation errors on form submission failure
    - Display success messages on successful operations
    - _Requirements: 2.3, 2.4, 2.5, 2.6, 8.6_
  
  - [ ]* 10.6 Write property test for validation feedback display
    - **Property 26: Validation Feedback Display**
    - **Validates: Requirements 8.4, 8.6**

- [x] 11. Implement settings handlers and UI
  - [x] 11.1 Create settings handler and template
    - Implement GET /settings handler to display SMTP settings
    - Create settings.html template with SMTP configuration fields
    - Mask password field in display (show asterisks)
    - _Requirements: 4.1_
  
  - [x] 11.2 Create settings update handler
    - Implement POST /settings handler to update SMTP settings
    - Validate settings before saving
    - Display validation errors or success message
    - _Requirements: 4.2, 4.3, 4.5_

- [x] 12. Create base template and styling
  - [x] 12.1 Create base HTML template
    - Implement base.html with navigation, header, footer
    - Add Turkish language labels throughout
    - Include CSS and JavaScript references
    - _Requirements: 8.1_
  
  - [x] 12.2 Create CSS styling
    - Implement modern, clean CSS in static/css/style.css
    - Add responsive design for mobile, tablet, desktop
    - Style forms, buttons, tables, error messages
    - _Requirements: 8.2, 8.5_
  
  - [x] 12.3 Create client-side JavaScript
    - Implement form validation helpers in static/js/app.js
    - Add loading state indicators
    - Add confirmation dialogs for delete operations
    - _Requirements: 8.6_

- [x] 13. Implement background scheduler service
  - [x] 13.1 Create scheduler service
    - Implement SchedulerService that runs in a goroutine
    - Check for due reminders every 60 seconds
    - Query reminders where next_send_at <= current time
    - _Requirements: 5.1_
  
  - [ ]* 13.2 Write property test for due reminder detection
    - **Property 18: Due Reminder Detection**
    - **Validates: Requirements 5.1**
  
  - [x] 13.3 Implement reminder processing logic
    - Load SMTP settings for reminder's user
    - Send email to all recipients using EmailService
    - Update last_sent_at timestamp on success
    - Calculate and update next_send_at for recurring reminders
    - _Requirements: 5.2, 5.3, 5.4, 5.6_
  
  - [ ]* 13.4 Write property test for email sending to all recipients
    - **Property 19: Email Sending to All Recipients**
    - **Validates: Requirements 5.2, 5.3**
  
  - [ ]* 13.5 Write property test for last-sent timestamp update
    - **Property 20: Last-Sent Timestamp Update**
    - **Validates: Requirements 5.4**
  
  - [ ]* 13.6 Write property test for next send time recalculation
    - **Property 21: Next Send Time Recalculation**
    - **Validates: Requirements 5.6**
  
  - [x] 13.7 Implement error handling and retry logic
    - Log errors with structured logging (timestamp, reminder ID, error details)
    - Implement exponential backoff retry (3 attempts: 1 min, 5 min, 15 min)
    - Mark reminders as failed after max retries
    - _Requirements: 5.5, 9.1_
  
  - [ ]* 13.8 Write unit tests for retry logic
    - Test exponential backoff timing
    - Test max retry limit
    - Test error logging
    - _Requirements: 5.5, 9.1_

- [x] 14. Implement logging infrastructure
  - [x] 14.1 Create structured logging utility
    - Implement logger with levels (INFO, WARN, ERROR)
    - Log authentication attempts (successful and failed)
    - Log email sending results
    - Log database errors
    - Mask sensitive data in logs (passwords, session IDs)
    - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5_
  
  - [ ]* 14.2 Write property test for comprehensive error logging
    - **Property 27: Comprehensive Error Logging**
    - **Validates: Requirements 9.1, 9.2, 9.4, 9.5**

- [x] 15. Implement configuration management
  - [x] 15.1 Create configuration loader
    - Load configuration from environment variables
    - Support: DATABASE_PATH, ENCRYPTION_KEY, SERVER_PORT, SESSION_DURATION
    - Provide sensible defaults
    - _Requirements: 10.4_
  
  - [x] 15.2 Wire all components in main.go
    - Initialize database connection
    - Initialize repository
    - Initialize all services (auth, reminder, settings, email, scheduler)
    - Initialize handlers with dependency injection
    - Set up HTTP routes with authentication middleware
    - Start scheduler in background goroutine
    - Start HTTP server
    - _Requirements: 10.5_

- [x] 16. Create initial user setup
  - [x] 16.1 Implement user creation utility
    - Create a command-line utility or initialization function to create first user
    - Hash password and store in database
    - Document how to create additional users
    - _Requirements: 1.2, 1.5_

- [-] 17. Write integration tests
  - [-]* 17.1 Write end-to-end reminder creation and sending test
    - Test complete flow: login → create reminder → wait for due time → verify email sent
    - Use in-memory database and mock SMTP
    - _Requirements: 1.2, 2.3, 5.2_
  
  - [ ]* 17.2 Write SMTP settings update and usage test
    - Test: update SMTP settings → create reminder → verify new settings used
    - _Requirements: 4.3, 5.3_
  
  - [ ]* 17.3 Write property test for data persistence across restarts
    - **Property 23: Data Persistence Across Restarts**
    - **Validates: Requirements 7.5**

- [x] 18. Final checkpoint - Ensure all tests pass
  - Run all unit tests and property tests
  - Verify test coverage meets requirements
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Property tests validate universal correctness properties with 100+ iterations
- Unit tests validate specific examples and edge cases
- The scheduler runs as a background goroutine and should be gracefully stopped on application shutdown
- Use in-memory SQLite (`:memory:`) for faster test execution
- Mock the SMTP client in tests to avoid sending real emails
- Use dependency injection throughout for better testability
