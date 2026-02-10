package services

import (
	"email-reminder-system/database"
	"email-reminder-system/models"
	"fmt"
	"regexp"
	"strings"
)

// SettingsService handles SMTP settings management
type SettingsService struct {
	repo      database.Repository
	encryptor *EncryptionService
}

// NewSettingsService creates a new settings service
func NewSettingsService(repo database.Repository, encryptor *EncryptionService) *SettingsService {
	return &SettingsService{
		repo:      repo,
		encryptor: encryptor,
	}
}

// ValidateSMTPSettings validates SMTP settings
func (s *SettingsService) ValidateSMTPSettings(settings *models.SMTPSettings) error {
	var errors []string

	// Validate required fields
	if strings.TrimSpace(settings.Host) == "" {
		errors = append(errors, "host is required")
	}

	if settings.Port < 1 || settings.Port > 65535 {
		errors = append(errors, "port must be between 1 and 65535")
	}

	if strings.TrimSpace(settings.Username) == "" {
		errors = append(errors, "username is required")
	}

	if strings.TrimSpace(settings.Password) == "" {
		errors = append(errors, "password is required")
	}

	if strings.TrimSpace(settings.FromEmail) == "" {
		errors = append(errors, "from_email is required")
	} else if !isValidEmail(settings.FromEmail) {
		errors = append(errors, "from_email must be a valid email address")
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// GetSMTPSettings retrieves SMTP settings for a user
func (s *SettingsService) GetSMTPSettings(userID int64) (*models.SMTPSettings, error) {
	settings, err := s.repo.GetSMTPSettings(userID)
	if err != nil {
		return nil, err
	}

	// Decrypt password
	decryptedPassword, err := s.encryptor.Decrypt(settings.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt password: %w", err)
	}
	settings.Password = decryptedPassword

	return settings, nil
}

// UpsertSMTPSettings creates or updates SMTP settings for a user
func (s *SettingsService) UpsertSMTPSettings(settings *models.SMTPSettings) error {
	// Validate settings
	if err := s.ValidateSMTPSettings(settings); err != nil {
		return err
	}

	// Encrypt password before storage
	encryptedPassword, err := s.encryptor.Encrypt(settings.Password)
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}

	// Create a copy to avoid modifying the original
	settingsToStore := *settings
	settingsToStore.Password = encryptedPassword

	// Store in database
	if err := s.repo.UpsertSMTPSettings(&settingsToStore); err != nil {
		return err
	}

	// Update the original with the stored ID and timestamp
	settings.ID = settingsToStore.ID
	settings.UpdatedAt = settingsToStore.UpdatedAt

	return nil
}

// isValidEmail validates email address format using RFC 5322 regex
func isValidEmail(email string) bool {
	// Simplified RFC 5322 email regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
