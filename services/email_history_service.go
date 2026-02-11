package services

import (
	"fmt"
	"time"

	"email-reminder-system/database"
	"email-reminder-system/models"
)

// EmailHistoryService handles email logging and history operations
type EmailHistoryService struct {
	repo   database.Repository
	logger *Logger
}

// NewEmailHistoryService creates a new email history service
func NewEmailHistoryService(repo database.Repository, logger *Logger) *EmailHistoryService {
	return &EmailHistoryService{
		repo:   repo,
		logger: logger,
	}
}

// LogEmailAttempt creates an email log record with "pending" status before sending
func (s *EmailHistoryService) LogEmailAttempt(
	reminderID *int64,
	userID int64,
	recipients []string,
	subject string,
	emailContent string,
) (int64, error) {
	log := &models.EmailLog{
		ReminderID:   reminderID,
		UserID:       userID,
		Recipients:   recipients,
		Subject:      subject,
		EmailContent: emailContent,
		Status:       "pending",
		SentAt:       time.Now(),
	}

	logID, err := s.repo.CreateEmailLog(log)
	if err != nil {
		s.logger.Warn("Failed to log email attempt", map[string]interface{}{
			"user_id":        userID,
			"reminder_id":    reminderID,
			"recipient_count": len(recipients),
			"error":          err.Error(),
		})
		return 0, err
	}

	s.logger.Info("Email attempt logged", map[string]interface{}{
		"log_id":         logID,
		"user_id":        userID,
		"reminder_id":    reminderID,
		"recipient_count": len(recipients),
	})

	return logID, nil
}

// LogEmailSuccess updates an email log to "success" status after successful send
func (s *EmailHistoryService) LogEmailSuccess(logID int64) error {
	err := s.repo.UpdateEmailLogStatus(logID, "success", nil, time.Now())
	if err != nil {
		s.logger.Warn("Failed to log email success", map[string]interface{}{
			"log_id": logID,
			"error":  err.Error(),
		})
		return err
	}

	s.logger.Info("Email success logged", map[string]interface{}{
		"log_id": logID,
	})

	return nil
}

// LogEmailFailure updates an email log to "failure" status with error message
func (s *EmailHistoryService) LogEmailFailure(logID int64, errorMsg string) error {
	// Truncate error message to 1000 characters if needed
	truncatedMsg := errorMsg
	if len(errorMsg) > 1000 {
		truncatedMsg = errorMsg[:1000] + "..."
	}

	err := s.repo.UpdateEmailLogStatus(logID, "failure", &truncatedMsg, time.Now())
	if err != nil {
		s.logger.Warn("Failed to log email failure", map[string]interface{}{
			"log_id": logID,
			"error":  err.Error(),
		})
		return err
	}

	s.logger.Info("Email failure logged", map[string]interface{}{
		"log_id":        logID,
		"error_message": truncatedMsg,
	})

	return nil
}

// GetEmailHistory retrieves paginated email history with filters
func (s *EmailHistoryService) GetEmailHistory(
	filter *models.EmailHistoryFilter,
) (*models.EmailHistoryResult, error) {
	// Validate and set defaults for filter parameters
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}

	result, err := s.repo.GetEmailLogs(filter)
	if err != nil {
		s.logger.Error("Failed to get email history", map[string]interface{}{
			"user_id": filter.UserID,
			"page":    filter.Page,
			"error":   err.Error(),
		})
		return nil, fmt.Errorf("failed to get email history: %w", err)
	}

	s.logger.Info("Email history retrieved", map[string]interface{}{
		"user_id":     filter.UserID,
		"page":        result.Page,
		"total_count": result.TotalCount,
	})

	return result, nil
}

// GetEmailLogDetail retrieves a single email log with user authorization
func (s *EmailHistoryService) GetEmailLogDetail(
	logID int64,
	userID int64,
) (*models.EmailLog, error) {
	log, err := s.repo.GetEmailLog(logID, userID)
	if err != nil {
		s.logger.Warn("Failed to get email log detail", map[string]interface{}{
			"log_id":  logID,
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, fmt.Errorf("failed to get email log detail: %w", err)
	}

	s.logger.Info("Email log detail retrieved", map[string]interface{}{
		"log_id":  logID,
		"user_id": userID,
	})

	return log, nil
}

// GetUserReminders retrieves all reminders for a user (for filter dropdown)
func (s *EmailHistoryService) GetUserReminders(userID int64) ([]*models.Reminder, error) {
	reminders, err := s.repo.GetRemindersByUserID(userID)
	if err != nil {
		s.logger.Error("Failed to get user reminders", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, fmt.Errorf("failed to get user reminders: %w", err)
	}

	return reminders, nil
}
