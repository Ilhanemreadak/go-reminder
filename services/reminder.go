package services

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"email-reminder-system/database"
	"email-reminder-system/models"
)

// ReminderService defines the interface for reminder operations
type ReminderService interface {
	CreateReminder(userID int64, reminder *models.Reminder) error
	GetReminder(id, userID int64) (*models.Reminder, error)
	ListReminders(userID int64) ([]*models.Reminder, error)
	UpdateReminder(reminder *models.Reminder, userID int64) error
	DeleteReminder(id, userID int64) error
	CalculateNextSendTime(reminder *models.Reminder) (time.Time, error)
	ValidateReminder(reminder *models.Reminder) error
}

// reminderService implements ReminderService interface
type reminderService struct {
	repo database.Repository
}

// NewReminderService creates a new reminder service
func NewReminderService(repo database.Repository) ReminderService {
	return &reminderService{
		repo: repo,
	}
}

// RFC 5322 compliant email regex (simplified but practical)
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

// Time format regex for HH:MM in 24-hour format
var timeFormatRegex = regexp.MustCompile(`^([01][0-9]|2[0-3]):([0-5][0-9])$`)

// Date format regex for YYYY-MM-DD
var dateFormatRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// ValidateReminder validates all fields of a reminder
func (s *reminderService) ValidateReminder(reminder *models.Reminder) error {
	var errors []string

	// Validate title
	if strings.TrimSpace(reminder.Title) == "" {
		errors = append(errors, "title is required")
	}

	// Validate recipients
	if len(reminder.Recipients) == 0 {
		errors = append(errors, "at least one recipient is required")
	} else {
		invalidEmails := []string{}
		for _, email := range reminder.Recipients {
			if !s.validateEmailAddress(email) {
				invalidEmails = append(invalidEmails, email)
			}
		}
		if len(invalidEmails) > 0 {
			errors = append(errors, fmt.Sprintf("invalid email addresses: %s", strings.Join(invalidEmails, ", ")))
		}
	}

	// Validate email content
	if strings.TrimSpace(reminder.EmailContent) == "" {
		errors = append(errors, "email content is required")
	}

	// Validate schedule type
	validScheduleTypes := map[string]bool{
		"once":    true,
		"daily":   true,
		"weekly":  true,
		"monthly": true,
		"custom":  true,
	}
	if !validScheduleTypes[reminder.ScheduleType] {
		errors = append(errors, "schedule type must be one of: once, daily, weekly, monthly, custom")
	}

	// Validate time of day format
	if !s.validateTimeFormat(reminder.TimeOfDay) {
		errors = append(errors, "time of day must be in HH:MM format (24-hour)")
	}

	// Validate schedule-specific fields
	switch reminder.ScheduleType {
	case "once":
		// Validate schedule date
		if !dateFormatRegex.MatchString(reminder.ScheduleDate) {
			errors = append(errors, "schedule date must be in YYYY-MM-DD format")
		} else {
			// Parse the date to verify it's valid
			_, err := time.Parse("2006-01-02", reminder.ScheduleDate)
			if err != nil {
				errors = append(errors, "invalid schedule date")
			}
		}
	case "daily":
		// Daily only requires time of day (already validated above)
	case "weekly":
		if reminder.DayOfWeek < 0 || reminder.DayOfWeek > 6 {
			errors = append(errors, "day of week must be between 0 (Sunday) and 6 (Saturday)")
		}
	case "monthly":
		if reminder.DayOfMonth < 1 || reminder.DayOfMonth > 31 {
			errors = append(errors, "day of month must be between 1 and 31")
		}
	case "custom":
		if reminder.IntervalDays < 1 {
			errors = append(errors, "interval days must be at least 1 for custom schedules")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// validateEmailAddress validates an email address using RFC 5322 regex
func (s *reminderService) validateEmailAddress(email string) bool {
	email = strings.TrimSpace(email)
	if len(email) == 0 || len(email) > 254 {
		return false
	}
	return emailRegex.MatchString(email)
}

// validateTimeFormat validates time format as HH:MM in 24-hour format
func (s *reminderService) validateTimeFormat(timeStr string) bool {
	return timeFormatRegex.MatchString(timeStr)
}

// CalculateNextSendTime calculates the next send time based on schedule configuration
func (s *reminderService) CalculateNextSendTime(reminder *models.Reminder) (time.Time, error) {
	now := time.Now()

	// Parse time of day
	parts := strings.Split(reminder.TimeOfDay, ":")
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("invalid time format")
	}

	var hour, minute int
	fmt.Sscanf(parts[0], "%d", &hour)
	fmt.Sscanf(parts[1], "%d", &minute)

	var nextSend time.Time

	switch reminder.ScheduleType {
	case "once":
		// One-time reminder: use the specific schedule date
		scheduleDate, err := time.Parse("2006-01-02", reminder.ScheduleDate)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid schedule date: %w", err)
		}
		nextSend = time.Date(scheduleDate.Year(), scheduleDate.Month(), scheduleDate.Day(), hour, minute, 0, 0, now.Location())

	case "daily":
		// Calculate next daily occurrence
		nextSend = time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
		if nextSend.Before(now) || nextSend.Equal(now) {
			// If time has passed today, schedule for tomorrow
			nextSend = nextSend.AddDate(0, 0, 1)
		}

	case "weekly":
		// Calculate next weekly occurrence
		currentWeekday := int(now.Weekday())
		daysUntilTarget := (reminder.DayOfWeek - currentWeekday + 7) % 7

		nextSend = time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
		nextSend = nextSend.AddDate(0, 0, daysUntilTarget)

		// If it's the same day but time has passed, schedule for next week
		if daysUntilTarget == 0 && (nextSend.Before(now) || nextSend.Equal(now)) {
			nextSend = nextSend.AddDate(0, 0, 7)
		}

	case "monthly":
		// Calculate next monthly occurrence
		nextSend = time.Date(now.Year(), now.Month(), reminder.DayOfMonth, hour, minute, 0, 0, now.Location())

		// Handle invalid dates (e.g., day 31 in February)
		if nextSend.Day() != reminder.DayOfMonth {
			// Move to next month and try again
			nextSend = time.Date(now.Year(), now.Month()+1, reminder.DayOfMonth, hour, minute, 0, 0, now.Location())
		}

		// If time has passed this month, schedule for next month
		if nextSend.Before(now) || nextSend.Equal(now) {
			nextSend = time.Date(now.Year(), now.Month()+1, reminder.DayOfMonth, hour, minute, 0, 0, now.Location())
			// Handle invalid dates again
			if nextSend.Day() != reminder.DayOfMonth {
				nextSend = time.Date(now.Year(), now.Month()+2, reminder.DayOfMonth, hour, minute, 0, 0, now.Location())
			}
		}

	case "custom":
		// Calculate next custom interval occurrence
		if reminder.LastSentAt != nil {
			// Use last sent time as base
			nextSend = time.Date(reminder.LastSentAt.Year(), reminder.LastSentAt.Month(),
				reminder.LastSentAt.Day(), hour, minute, 0, 0, reminder.LastSentAt.Location())
			nextSend = nextSend.AddDate(0, 0, reminder.IntervalDays)
		} else {
			// Use current time as base for first occurrence
			nextSend = time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
			if nextSend.Before(now) || nextSend.Equal(now) {
				// If time has passed today, add interval days
				nextSend = nextSend.AddDate(0, 0, reminder.IntervalDays)
			}
		}

	default:
		return time.Time{}, fmt.Errorf("invalid schedule type: %s", reminder.ScheduleType)
	}

	return nextSend, nil
}

// CreateReminder creates a new reminder with validation and next_send_at calculation
func (s *reminderService) CreateReminder(userID int64, reminder *models.Reminder) error {
	// Set user ID
	reminder.UserID = userID
	reminder.IsActive = true

	// Validate reminder
	if err := s.ValidateReminder(reminder); err != nil {
		return err
	}

	// Calculate next send time
	nextSend, err := s.CalculateNextSendTime(reminder)
	if err != nil {
		return fmt.Errorf("failed to calculate next send time: %w", err)
	}
	reminder.NextSendAt = nextSend

	// Log the calculated next send time
	logger := GetLogger()
	logger.Info("Creating reminder", map[string]interface{}{
		"title":         reminder.Title,
		"schedule_type": reminder.ScheduleType,
		"time_of_day":   reminder.TimeOfDay,
		"next_send_at":  nextSend.Format("2006-01-02 15:04:05"),
	})

	// Create in database
	if err := s.repo.CreateReminder(reminder); err != nil {
		return fmt.Errorf("failed to create reminder: %w", err)
	}

	return nil
}

// GetReminder retrieves a reminder by ID with user authorization
func (s *reminderService) GetReminder(id, userID int64) (*models.Reminder, error) {
	reminder, err := s.repo.GetReminder(id)
	if err != nil {
		return nil, err
	}

	// Check user authorization
	if reminder.UserID != userID {
		return nil, fmt.Errorf("unauthorized: reminder does not belong to user")
	}

	return reminder, nil
}

// ListReminders retrieves all reminders for a user
func (s *reminderService) ListReminders(userID int64) ([]*models.Reminder, error) {
	reminders, err := s.repo.GetRemindersByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list reminders: %w", err)
	}
	return reminders, nil
}

// UpdateReminder updates an existing reminder with validation and user authorization
func (s *reminderService) UpdateReminder(reminder *models.Reminder, userID int64) error {
	// Check if reminder exists and belongs to user
	existing, err := s.repo.GetReminder(reminder.ID)
	if err != nil {
		return err
	}

	if existing.UserID != userID {
		return fmt.Errorf("unauthorized: reminder does not belong to user")
	}

	// Validate reminder
	reminder.UserID = userID
	if err := s.ValidateReminder(reminder); err != nil {
		return err
	}

	// Recalculate next send time
	nextSend, err := s.CalculateNextSendTime(reminder)
	if err != nil {
		return fmt.Errorf("failed to calculate next send time: %w", err)
	}
	reminder.NextSendAt = nextSend

	// Update in database
	if err := s.repo.UpdateReminder(reminder); err != nil {
		return fmt.Errorf("failed to update reminder: %w", err)
	}

	return nil
}

// DeleteReminder deletes a reminder with user authorization
func (s *reminderService) DeleteReminder(id, userID int64) error {
	// Check if reminder exists and belongs to user
	reminder, err := s.repo.GetReminder(id)
	if err != nil {
		return err
	}

	if reminder.UserID != userID {
		return fmt.Errorf("unauthorized: reminder does not belong to user")
	}

	// Delete from database
	if err := s.repo.DeleteReminder(id); err != nil {
		return fmt.Errorf("failed to delete reminder: %w", err)
	}

	return nil
}
