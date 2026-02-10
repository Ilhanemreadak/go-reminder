package models

import "time"

// Reminder represents a scheduled email reminder
type Reminder struct {
	ID           int64
	UserID       int64
	Title        string
	Recipients   []string // Email addresses
	EmailContent string
	ScheduleType string // "daily", "weekly", "monthly", "custom"
	IntervalDays int    // For custom schedules
	DayOfWeek    int    // 0-6 for weekly (0=Sunday)
	DayOfMonth   int    // 1-31 for monthly
	TimeOfDay    string // HH:MM format
	LastSentAt   *time.Time
	NextSendAt   time.Time
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
