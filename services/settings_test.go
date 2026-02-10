package services

import (
	"email-reminder-system/models"
	"testing"
)

func TestValidateSMTPSettings(t *testing.T) {
	service := &SettingsService{}

	tests := []struct {
		name        string
		settings    *models.SMTPSettings
		expectError bool
	}{
		{
			name: "valid settings",
			settings: &models.SMTPSettings{
				Host:      "smtp.example.com",
				Port:      587,
				Username:  "user@example.com",
				Password:  "password123",
				FromEmail: "sender@example.com",
				UseTLS:    true,
			},
			expectError: false,
		},
		{
			name: "missing host",
			settings: &models.SMTPSettings{
				Port:      587,
				Username:  "user@example.com",
				Password:  "password123",
				FromEmail: "sender@example.com",
			},
			expectError: true,
		},
		{
			name: "invalid port - too low",
			settings: &models.SMTPSettings{
				Host:      "smtp.example.com",
				Port:      0,
				Username:  "user@example.com",
				Password:  "password123",
				FromEmail: "sender@example.com",
			},
			expectError: true,
		},
		{
			name: "invalid port - too high",
			settings: &models.SMTPSettings{
				Host:      "smtp.example.com",
				Port:      65536,
				Username:  "user@example.com",
				Password:  "password123",
				FromEmail: "sender@example.com",
			},
			expectError: true,
		},
		{
			name: "missing username",
			settings: &models.SMTPSettings{
				Host:      "smtp.example.com",
				Port:      587,
				Password:  "password123",
				FromEmail: "sender@example.com",
			},
			expectError: true,
		},
		{
			name: "missing password",
			settings: &models.SMTPSettings{
				Host:      "smtp.example.com",
				Port:      587,
				Username:  "user@example.com",
				FromEmail: "sender@example.com",
			},
			expectError: true,
		},
		{
			name: "missing from_email",
			settings: &models.SMTPSettings{
				Host:     "smtp.example.com",
				Port:     587,
				Username: "user@example.com",
				Password: "password123",
			},
			expectError: true,
		},
		{
			name: "invalid from_email format",
			settings: &models.SMTPSettings{
				Host:      "smtp.example.com",
				Port:      587,
				Username:  "user@example.com",
				Password:  "password123",
				FromEmail: "invalid-email",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateSMTPSettings(tt.settings)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
		})
	}
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"user@example.com", true},
		{"test.user@example.com", true},
		{"user+tag@example.co.uk", true},
		{"invalid-email", false},
		{"@example.com", false},
		{"user@", false},
		{"user", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := isValidEmail(tt.email)
			if result != tt.valid {
				t.Errorf("isValidEmail(%q) = %v, want %v", tt.email, result, tt.valid)
			}
		})
	}
}
