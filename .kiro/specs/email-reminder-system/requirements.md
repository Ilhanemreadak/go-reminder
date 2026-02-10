# Requirements Document

## Introduction

The Email Reminder System is a web application that enables users to create, manage, and automatically send scheduled email reminders. The system provides a Turkish-language interface for managing reminders with configurable schedules, SMTP settings, and secure authentication. Built with Go and SQLite, it offers a modern, clean UI for reminder management and automated background email delivery.

## Glossary

- **System**: The Email Reminder System web application
- **User**: An authenticated person who manages reminders and settings
- **Reminder**: A scheduled email notification with recipients, content, and timing configuration
- **SMTP_Settings**: Configuration parameters for the email server (host, port, credentials)
- **Schedule**: The timing configuration that determines when a reminder email is sent
- **Session**: An authenticated user's active connection to the system
- **Background_Job**: The automated process that sends scheduled emails

## Requirements

### Requirement 1: User Authentication

**User Story:** As a system administrator, I want secure password-protected access to the application, so that only authorized users can manage reminders and settings.

#### Acceptance Criteria

1. WHEN a user attempts to access any application page without authentication, THE System SHALL redirect them to the login page
2. WHEN a user submits valid credentials, THE System SHALL create a Session and grant access to protected pages
3. WHEN a user submits invalid credentials, THE System SHALL reject the login attempt and display an error message
4. WHEN a Session expires or is invalidated, THE System SHALL require re-authentication before allowing further access
5. THE System SHALL store authentication credentials securely using industry-standard hashing

### Requirement 2: Reminder Management

**User Story:** As a user, I want to create, view, edit, and delete email reminders, so that I can manage my scheduled notifications effectively.

#### Acceptance Criteria

1. WHEN a user accesses the main page, THE System SHALL display a list of all existing Reminders with their titles and schedules
2. WHEN a user clicks "Hatırlatıcı Ekle" (Add Reminder), THE System SHALL display a creation form
3. WHEN a user submits a valid reminder form, THE System SHALL create a new Reminder and store it in the database
4. WHEN a user submits an invalid reminder form, THE System SHALL reject the submission and display validation errors
5. WHEN a user edits an existing Reminder, THE System SHALL update the Reminder in the database and preserve data integrity
6. WHEN a user deletes a Reminder, THE System SHALL remove it from the database and cancel any pending scheduled emails

### Requirement 3: Reminder Data Collection

**User Story:** As a user, I want to specify all necessary details for a reminder, so that the system can send the correct emails at the right time.

#### Acceptance Criteria

1. WHEN creating or editing a Reminder, THE System SHALL require a title field
2. WHEN creating or editing a Reminder, THE System SHALL require at least one valid email address in the recipient list
3. WHEN creating or editing a Reminder, THE System SHALL require email content or template
4. WHEN creating or editing a Reminder, THE System SHALL require a schedule type selection (daily, weekly, monthly, custom)
5. WHEN creating or editing a Reminder, THE System SHALL require interval settings appropriate to the schedule type
6. WHEN a user provides invalid email addresses, THE System SHALL reject the submission and indicate which addresses are invalid

### Requirement 4: SMTP Configuration

**User Story:** As a user, I want to configure SMTP settings, so that the system can send emails through my email server.

#### Acceptance Criteria

1. WHEN a user accesses the settings page, THE System SHALL display current SMTP_Settings
2. WHEN a user updates SMTP_Settings, THE System SHALL validate the configuration parameters
3. WHEN a user saves valid SMTP_Settings, THE System SHALL store them securely in the database
4. WHEN SMTP_Settings contain credentials, THE System SHALL encrypt sensitive data before storage
5. WHEN SMTP_Settings are incomplete or invalid, THE System SHALL prevent saving and display validation errors

### Requirement 5: Scheduled Email Sending

**User Story:** As a user, I want the system to automatically send reminder emails according to configured schedules, so that I don't have to manually trigger each reminder.

#### Acceptance Criteria

1. THE Background_Job SHALL continuously monitor for Reminders that are due to be sent
2. WHEN a Reminder's schedule indicates it is time to send, THE Background_Job SHALL retrieve the Reminder details and send emails to all recipients
3. WHEN sending an email, THE Background_Job SHALL use the configured SMTP_Settings
4. WHEN an email is successfully sent, THE Background_Job SHALL update the Reminder's last-sent timestamp
5. WHEN an email fails to send, THE Background_Job SHALL log the error and retry according to a retry policy
6. WHEN a Reminder has a recurring schedule, THE Background_Job SHALL calculate and schedule the next send time after successful delivery

### Requirement 6: Schedule Configuration

**User Story:** As a user, I want to configure flexible schedules for reminders, so that I can set up daily, weekly, or custom recurring notifications.

#### Acceptance Criteria

1. WHEN a schedule type is "daily", THE System SHALL require a time-of-day setting
2. WHEN a schedule type is "weekly", THE System SHALL require day-of-week selection and time-of-day setting
3. WHEN a schedule type is "monthly", THE System SHALL require day-of-month selection and time-of-day setting
4. WHEN a schedule type is "custom", THE System SHALL require an interval in days and time-of-day setting
5. THE System SHALL validate that all time settings are in valid 24-hour format
6. THE System SHALL calculate the next scheduled send time based on the current time and schedule configuration

### Requirement 7: Data Persistence

**User Story:** As a user, I want all reminders and settings to be stored reliably, so that my data persists across application restarts.

#### Acceptance Criteria

1. THE System SHALL store all Reminder data in a SQLite database
2. THE System SHALL store SMTP_Settings in the SQLite database
3. THE System SHALL store Session data in the SQLite database or secure session store
4. WHEN the application starts, THE System SHALL initialize the database schema if it does not exist
5. WHEN the application starts, THE System SHALL load existing Reminders and SMTP_Settings from the database
6. THE System SHALL maintain referential integrity for all database relationships

### Requirement 8: User Interface

**User Story:** As a user, I want a modern, clean, and responsive interface, so that I can easily manage reminders on any device.

#### Acceptance Criteria

1. THE System SHALL display a Turkish-language interface with "Hatırlatıcılar" as the main page title
2. THE System SHALL render properly on desktop, tablet, and mobile screen sizes
3. WHEN displaying the reminder list, THE System SHALL show key information (title, schedule, next send time) in a clear format
4. WHEN displaying forms, THE System SHALL provide clear labels and input validation feedback
5. THE System SHALL use modern CSS styling for a professional appearance
6. THE System SHALL provide visual feedback for user actions (loading states, success messages, error messages)

### Requirement 9: Error Handling and Logging

**User Story:** As a system administrator, I want comprehensive error handling and logging, so that I can troubleshoot issues and monitor system health.

#### Acceptance Criteria

1. WHEN an error occurs during email sending, THE System SHALL log the error with timestamp, reminder ID, and error details
2. WHEN an error occurs during database operations, THE System SHALL log the error and return appropriate error messages to the user
3. WHEN SMTP connection fails, THE System SHALL log the connection error and notify the user
4. THE System SHALL log all authentication attempts (successful and failed)
5. THE System SHALL provide structured logging with appropriate log levels (INFO, WARN, ERROR)

### Requirement 10: Project Structure

**User Story:** As a developer, I want a professional folder structure and file organization, so that the codebase is maintainable and follows Go best practices.

#### Acceptance Criteria

1. THE System SHALL organize code into logical packages (handlers, models, services, database)
2. THE System SHALL separate HTML templates into a dedicated templates directory
3. THE System SHALL separate static assets (CSS, JavaScript) into a dedicated static directory
4. THE System SHALL use a configuration file or environment variables for application settings
5. THE System SHALL include a main.go entry point that initializes all components
