package services

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	// INFO level for informational messages
	INFO LogLevel = iota
	// WARN level for warning messages
	WARN
	// ERROR level for error messages
	ERROR
)

// String returns the string representation of a log level
func (l LogLevel) String() string {
	switch l {
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger provides structured logging with sensitive data masking
type Logger struct {
	logger  *log.Logger
	logFile *os.File
}

// NewLogger creates a new logger instance that writes to both console and file
func NewLogger() *Logger {
	// Create logs directory if it doesn't exist
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("Failed to create logs directory: %v", err)
		return &Logger{
			logger: log.New(os.Stdout, "", 0),
		}
	}

	// Create log file with date in filename
	logFileName := filepath.Join(logDir, fmt.Sprintf("app-%s.log", time.Now().Format("2006-01-02")))
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		return &Logger{
			logger: log.New(os.Stdout, "", 0),
		}
	}

	// Write to both console and file
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	return &Logger{
		logger:  log.New(multiWriter, "", 0),
		logFile: logFile,
	}
}

// Close closes the log file
func (l *Logger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

// log writes a log message with the specified level and context
func (l *Logger) log(level LogLevel, message string, context map[string]interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	
	// Build log message
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[%s] %s: %s", timestamp, level.String(), message))
	
	// Add context fields if provided
	if len(context) > 0 {
		sb.WriteString(" |")
		for key, value := range context {
			maskedValue := l.maskSensitiveData(key, value)
			sb.WriteString(fmt.Sprintf(" %s=%v", key, maskedValue))
		}
	}
	
	l.logger.Println(sb.String())
}

// maskSensitiveData masks sensitive information in log output
func (l *Logger) maskSensitiveData(key string, value interface{}) interface{} {
	keyLower := strings.ToLower(key)
	
	// Mask passwords
	if strings.Contains(keyLower, "password") {
		return "********"
	}
	
	// Mask session IDs (show only first 8 characters)
	if strings.Contains(keyLower, "session") {
		if strValue, ok := value.(string); ok && len(strValue) > 8 {
			return strValue[:8] + "..."
		}
	}
	
	// Mask encryption keys
	if strings.Contains(keyLower, "key") || strings.Contains(keyLower, "secret") {
		return "********"
	}
	
	return value
}

// Info logs an informational message
func (l *Logger) Info(message string, context ...map[string]interface{}) {
	ctx := make(map[string]interface{})
	if len(context) > 0 {
		ctx = context[0]
	}
	l.log(INFO, message, ctx)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, context ...map[string]interface{}) {
	ctx := make(map[string]interface{})
	if len(context) > 0 {
		ctx = context[0]
	}
	l.log(WARN, message, ctx)
}

// Error logs an error message
func (l *Logger) Error(message string, context ...map[string]interface{}) {
	ctx := make(map[string]interface{})
	if len(context) > 0 {
		ctx = context[0]
	}
	l.log(ERROR, message, ctx)
}

// LogAuthAttempt logs an authentication attempt
func (l *Logger) LogAuthAttempt(username string, success bool, err error) {
	context := map[string]interface{}{
		"username": username,
		"success":  success,
	}
	
	if success {
		l.Info("Authentication successful", context)
	} else {
		if err != nil {
			context["error"] = err.Error()
		}
		l.Warn("Authentication failed", context)
	}
}

// LogEmailSent logs an email sending result
func (l *Logger) LogEmailSent(reminderID int64, recipients []string, success bool, err error) {
	context := map[string]interface{}{
		"reminder_id":      reminderID,
		"recipient_count":  len(recipients),
		"success":          success,
	}
	
	if success {
		l.Info("Email sent successfully", context)
	} else {
		if err != nil {
			context["error"] = err.Error()
		}
		l.Error("Email sending failed", context)
	}
}

// LogDatabaseError logs a database operation error
func (l *Logger) LogDatabaseError(operation string, err error, context ...map[string]interface{}) {
	ctx := map[string]interface{}{
		"operation": operation,
		"error":     err.Error(),
	}
	
	// Merge additional context if provided
	if len(context) > 0 {
		for key, value := range context[0] {
			ctx[key] = value
		}
	}
	
	l.Error("Database operation failed", ctx)
}

// LogSchedulerEvent logs scheduler-related events
func (l *Logger) LogSchedulerEvent(event string, context ...map[string]interface{}) {
	ctx := make(map[string]interface{})
	if len(context) > 0 {
		ctx = context[0]
	}
	l.Info(event, ctx)
}

// LogReminderProcessing logs reminder processing events
func (l *Logger) LogReminderProcessing(reminderID int64, title string, status string, context ...map[string]interface{}) {
	ctx := map[string]interface{}{
		"reminder_id": reminderID,
		"title":       title,
		"status":      status,
	}
	
	// Merge additional context if provided
	if len(context) > 0 {
		for key, value := range context[0] {
			ctx[key] = value
		}
	}
	
	l.Info("Reminder processing", ctx)
}

// Global logger instance
var globalLogger = NewLogger()

// GetLogger returns the global logger instance
func GetLogger() *Logger {
	return globalLogger
}
