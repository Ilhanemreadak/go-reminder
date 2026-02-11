package services

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"email-reminder-system/database"
	"email-reminder-system/models"
)

// SchedulerService handles background processing of due reminders
type SchedulerService struct {
	repo           database.Repository
	emailService   *EmailService
	reminderService ReminderService
	encryptionService *EncryptionService
	logger         *Logger
	ticker         *time.Ticker
	stopChan       chan struct{}
	wg             sync.WaitGroup
	ctx            context.Context
	cancel         context.CancelFunc
}

// NewSchedulerService creates a new scheduler service
func NewSchedulerService(repo database.Repository, emailService *EmailService, 
	reminderService ReminderService, encryptionService *EncryptionService) *SchedulerService {
	ctx, cancel := context.WithCancel(context.Background())
	return &SchedulerService{
		repo:              repo,
		emailService:      emailService,
		reminderService:   reminderService,
		encryptionService: encryptionService,
		logger:            GetLogger(),
		stopChan:          make(chan struct{}),
		ctx:               ctx,
		cancel:            cancel,
	}
}

// Start begins the background scheduler goroutine
func (s *SchedulerService) Start() error {
	s.logger.LogSchedulerEvent("Starting scheduler service")
	
	// Create ticker that fires every 60 seconds
	s.ticker = time.NewTicker(60 * time.Second)
	
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		
		// Panic recovery for scheduler goroutine
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error("Scheduler panic recovered", map[string]interface{}{
					"panic": fmt.Sprintf("%v", r),
				})
			}
		}()
		
		for {
			select {
			case <-s.ticker.C:
				// Process due reminders with error handling
				func() {
					defer func() {
						if r := recover(); r != nil {
							s.logger.Error("Panic in ProcessDueReminders", map[string]interface{}{
								"panic": fmt.Sprintf("%v", r),
							})
						}
					}()
					
					if err := s.ProcessDueReminders(); err != nil {
						s.logger.Error("Failed to process due reminders", map[string]interface{}{
							"error": err.Error(),
						})
					}
				}()
			case <-s.stopChan:
				s.logger.LogSchedulerEvent("Scheduler service stopping")
				return
			case <-s.ctx.Done():
				s.logger.LogSchedulerEvent("Scheduler service context cancelled")
				return
			}
		}
	}()
	
	return nil
}

// Stop gracefully stops the scheduler service
func (s *SchedulerService) Stop() error {
	s.logger.LogSchedulerEvent("Stopping scheduler service")
	
	if s.ticker != nil {
		s.ticker.Stop()
	}
	
	s.cancel()
	close(s.stopChan)
	s.wg.Wait()
	
	s.logger.LogSchedulerEvent("Scheduler service stopped")
	return nil
}

// ProcessDueReminders checks for and processes all due reminders
func (s *SchedulerService) ProcessDueReminders() error {
	now := time.Now()
	
	s.logger.Info("Checking for due reminders", map[string]interface{}{
		"current_time": now.Format("2006-01-02 15:04:05"),
	})
	
	// Query reminders where next_send_at <= current time
	dueReminders, err := s.repo.GetDueReminders(now)
	if err != nil {
		s.logger.LogDatabaseError("GetDueReminders", err)
		return fmt.Errorf("failed to get due reminders: %w", err)
	}
	
	s.logger.Info("Due reminders check complete", map[string]interface{}{
		"count": len(dueReminders),
	})
	
	if len(dueReminders) == 0 {
		return nil
	}
	
	s.logger.LogSchedulerEvent("Found due reminders to process", map[string]interface{}{
		"count": len(dueReminders),
	})
	
	// Process each due reminder
	for _, reminder := range dueReminders {
		s.logger.Info("Processing reminder", map[string]interface{}{
			"reminder_id":  reminder.ID,
			"title":        reminder.Title,
			"next_send_at": reminder.NextSendAt.Format("2006-01-02 15:04:05"),
		})
		
		if err := s.processReminder(reminder); err != nil {
			s.logger.Error("Failed to process reminder", map[string]interface{}{
				"reminder_id": reminder.ID,
				"error":       err.Error(),
			})
		}
	}
	
	return nil
}

// processReminder processes a single reminder with retry logic
func (s *SchedulerService) processReminder(reminder *models.Reminder) error {
	s.logger.LogReminderProcessing(reminder.ID, reminder.Title, "processing")
	
	// Load SMTP settings for the reminder's user
	smtpSettings, err := s.repo.GetSMTPSettings(reminder.UserID)
	if err != nil {
		s.logger.LogDatabaseError("GetSMTPSettings", err, map[string]interface{}{
			"user_id":     reminder.UserID,
			"reminder_id": reminder.ID,
		})
		return fmt.Errorf("failed to load SMTP settings for user %d: %w", reminder.UserID, err)
	}
	
	// Decrypt SMTP password
	decryptedPassword, err := s.encryptionService.Decrypt(smtpSettings.Password)
	if err != nil {
		s.logger.Error("Failed to decrypt SMTP password", map[string]interface{}{
			"user_id":     reminder.UserID,
			"reminder_id": reminder.ID,
			"error":       err.Error(),
		})
		return fmt.Errorf("failed to decrypt SMTP password: %w", err)
	}
	smtpSettings.Password = decryptedPassword
	
	// Attempt to send email with retry logic
	var lastErr error
	retryDelays := []time.Duration{1 * time.Minute, 5 * time.Minute, 15 * time.Minute}
	
	for attempt := 0; attempt <= len(retryDelays); attempt++ {
		if attempt > 0 {
			// Wait before retry
			delay := retryDelays[attempt-1]
			s.logger.Warn("Retrying reminder processing", map[string]interface{}{
				"reminder_id": reminder.ID,
				"attempt":     attempt,
				"delay":       delay.String(),
			})
			
			select {
			case <-time.After(delay):
				// Continue with retry
			case <-s.ctx.Done():
				return fmt.Errorf("scheduler stopped during retry")
			}
		}
		
		// Send email to all recipients
		err = s.emailService.SendEmail(reminder.Recipients, reminder.Title, reminder.EmailContent, smtpSettings, &reminder.ID, reminder.UserID)
		if err == nil {
			// Success - update timestamps and schedule next send
			s.logger.LogEmailSent(reminder.ID, reminder.Recipients, true, nil)
			
			// Update last_sent_at timestamp
			now := time.Now()
			reminder.LastSentAt = &now
			
			// Calculate and update next_send_at for recurring reminders
			nextSend, err := s.reminderService.CalculateNextSendTime(reminder)
			if err != nil {
				s.logger.Error("Failed to calculate next send time", map[string]interface{}{
					"reminder_id": reminder.ID,
					"error":       err.Error(),
				})
				return fmt.Errorf("failed to calculate next send time: %w", err)
			}
			reminder.NextSendAt = nextSend
			
			// Update reminder in database
			if err := s.repo.UpdateReminder(reminder); err != nil {
				s.logger.LogDatabaseError("UpdateReminder", err, map[string]interface{}{
					"reminder_id": reminder.ID,
				})
				return fmt.Errorf("failed to update reminder: %w", err)
			}
			
			s.logger.LogReminderProcessing(reminder.ID, reminder.Title, "scheduled", map[string]interface{}{
				"next_send_at": nextSend.Format(time.RFC3339),
			})
			return nil
		}
		
		lastErr = err
		s.logger.LogEmailSent(reminder.ID, reminder.Recipients, false, err)
	}
	
	// All retries exhausted - mark reminder as failed
	s.logger.Error("All retry attempts exhausted for reminder", map[string]interface{}{
		"reminder_id": reminder.ID,
		"attempts":    len(retryDelays) + 1,
	})
	reminder.IsActive = false
	if err := s.repo.UpdateReminder(reminder); err != nil {
		s.logger.LogDatabaseError("UpdateReminder", err, map[string]interface{}{
			"reminder_id": reminder.ID,
			"action":      "mark_as_failed",
		})
	}
	
	return fmt.Errorf("failed to send reminder after %d attempts: %w", len(retryDelays)+1, lastErr)
}

// calculateExponentialBackoff calculates exponential backoff delay
func calculateExponentialBackoff(attempt int, baseDelay time.Duration) time.Duration {
	return time.Duration(math.Pow(2, float64(attempt))) * baseDelay
}
