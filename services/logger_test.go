package services

import (
	"bytes"
	"errors"
	"log"
	"strings"
	"testing"
)

func TestLoggerLevels(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	logger := &Logger{
		logger: log.New(&buf, "", 0),
	}

	tests := []struct {
		name     string
		logFunc  func()
		expected string
	}{
		{
			name: "Info level",
			logFunc: func() {
				logger.Info("Test info message")
			},
			expected: "INFO: Test info message",
		},
		{
			name: "Warn level",
			logFunc: func() {
				logger.Warn("Test warning message")
			},
			expected: "WARN: Test warning message",
		},
		{
			name: "Error level",
			logFunc: func() {
				logger.Error("Test error message")
			},
			expected: "ERROR: Test error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc()
			output := buf.String()
			if !strings.Contains(output, tt.expected) {
				t.Errorf("Expected log to contain %q, got %q", tt.expected, output)
			}
		})
	}
}

func TestLoggerContext(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		logger: log.New(&buf, "", 0),
	}

	logger.Info("Test with context", map[string]interface{}{
		"user_id": 123,
		"action":  "create",
	})

	output := buf.String()
	if !strings.Contains(output, "user_id=123") {
		t.Errorf("Expected log to contain user_id=123, got %q", output)
	}
	if !strings.Contains(output, "action=create") {
		t.Errorf("Expected log to contain action=create, got %q", output)
	}
}

func TestSensitiveDataMasking(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		logger: log.New(&buf, "", 0),
	}

	tests := []struct {
		name        string
		context     map[string]interface{}
		shouldMask  string
		shouldNotContain string
	}{
		{
			name: "Password masking",
			context: map[string]interface{}{
				"username": "testuser",
				"password": "secretpassword123",
			},
			shouldMask:       "password=********",
			shouldNotContain: "secretpassword123",
		},
		{
			name: "Session ID masking",
			context: map[string]interface{}{
				"session_id": "abcdef123456789012345678",
			},
			shouldMask:       "session_id=abcdef12...",
			shouldNotContain: "abcdef123456789012345678",
		},
		{
			name: "Encryption key masking",
			context: map[string]interface{}{
				"encryption_key": "my-secret-key-12345",
			},
			shouldMask:       "encryption_key=********",
			shouldNotContain: "my-secret-key-12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			logger.Info("Test message", tt.context)
			output := buf.String()

			if !strings.Contains(output, tt.shouldMask) {
				t.Errorf("Expected log to contain %q, got %q", tt.shouldMask, output)
			}
			if strings.Contains(output, tt.shouldNotContain) {
				t.Errorf("Expected log NOT to contain %q, but it does: %q", tt.shouldNotContain, output)
			}
		})
	}
}

func TestLogAuthAttempt(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		logger: log.New(&buf, "", 0),
	}

	tests := []struct {
		name     string
		username string
		success  bool
		err      error
		expected []string
	}{
		{
			name:     "Successful authentication",
			username: "testuser",
			success:  true,
			err:      nil,
			expected: []string{"INFO", "Authentication successful", "username=testuser", "success=true"},
		},
		{
			name:     "Failed authentication",
			username: "testuser",
			success:  false,
			err:      errors.New("invalid credentials"),
			expected: []string{"WARN", "Authentication failed", "username=testuser", "success=false", "error=invalid credentials"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			logger.LogAuthAttempt(tt.username, tt.success, tt.err)
			output := buf.String()

			for _, exp := range tt.expected {
				if !strings.Contains(output, exp) {
					t.Errorf("Expected log to contain %q, got %q", exp, output)
				}
			}
		})
	}
}

func TestLogEmailSent(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		logger: log.New(&buf, "", 0),
	}

	tests := []struct {
		name       string
		reminderID int64
		recipients []string
		success    bool
		err        error
		expected   []string
	}{
		{
			name:       "Successful email send",
			reminderID: 123,
			recipients: []string{"user1@example.com", "user2@example.com"},
			success:    true,
			err:        nil,
			expected:   []string{"INFO", "Email sent successfully", "reminder_id=123", "recipient_count=2", "success=true"},
		},
		{
			name:       "Failed email send",
			reminderID: 456,
			recipients: []string{"user@example.com"},
			success:    false,
			err:        errors.New("SMTP connection failed"),
			expected:   []string{"ERROR", "Email sending failed", "reminder_id=456", "recipient_count=1", "success=false", "error=SMTP connection failed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			logger.LogEmailSent(tt.reminderID, tt.recipients, tt.success, tt.err)
			output := buf.String()

			for _, exp := range tt.expected {
				if !strings.Contains(output, exp) {
					t.Errorf("Expected log to contain %q, got %q", exp, output)
				}
			}
		})
	}
}

func TestLogDatabaseError(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		logger: log.New(&buf, "", 0),
	}

	err := errors.New("connection timeout")
	logger.LogDatabaseError("GetReminder", err, map[string]interface{}{
		"reminder_id": 789,
	})

	output := buf.String()
	expected := []string{"ERROR", "Database operation failed", "operation=GetReminder", "error=connection timeout", "reminder_id=789"}

	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("Expected log to contain %q, got %q", exp, output)
		}
	}
}

func TestLogReminderProcessing(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		logger: log.New(&buf, "", 0),
	}

	logger.LogReminderProcessing(123, "Daily Standup", "processing", map[string]interface{}{
		"attempt": 1,
	})

	output := buf.String()
	expected := []string{"INFO", "Reminder processing", "reminder_id=123", "title=Daily Standup", "status=processing", "attempt=1"}

	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("Expected log to contain %q, got %q", exp, output)
		}
	}
}

func TestGetLogger(t *testing.T) {
	logger := GetLogger()
	if logger == nil {
		t.Error("Expected GetLogger to return a non-nil logger")
	}
}
