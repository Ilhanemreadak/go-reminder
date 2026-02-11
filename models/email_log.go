package models

import "time"

// EmailLog represents a logged email sending attempt
type EmailLog struct {
	ID           int64
	ReminderID   *int64    // Nullable - can be NULL if reminder deleted
	UserID       int64
	Recipients   []string  // Deserialized from JSON
	Subject      string
	EmailContent string
	Status       string    // "pending", "success", "failure"
	ErrorMessage *string   // Nullable - only for failures
	SentAt       time.Time
	CreatedAt    time.Time

	// Joined fields (not in database)
	ReminderTitle string // From JOIN with reminders table
}

// EmailHistoryFilter represents filter criteria for email history queries
type EmailHistoryFilter struct {
	UserID     int64
	ReminderID *int64
	Status     string     // "", "success", "failure"
	StartDate  *time.Time
	EndDate    *time.Time
	Page       int        // 1-indexed
	PageSize   int        // Default 20
}

// EmailHistoryResult represents paginated email history results
type EmailHistoryResult struct {
	Logs       []*EmailLog
	TotalCount int
	Page       int
	PageSize   int
	TotalPages int
}
