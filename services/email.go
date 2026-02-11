package services

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"regexp"
	"strings"

	"email-reminder-system/models"
)

// EmailService handles email sending operations
type EmailService struct {
	logger         *Logger
	historyService *EmailHistoryService
}

// NewEmailService creates a new email service
func NewEmailService(historyService *EmailHistoryService) *EmailService {
	return &EmailService{
		logger:         GetLogger(),
		historyService: historyService,
	}
}

// SendEmail sends an email to the specified recipients using SMTP settings
func (s *EmailService) SendEmail(to []string, subject, body string, settings *models.SMTPSettings, reminderID *int64, userID int64) error {
	// Log email attempt before sending
	var logID int64
	if s.historyService != nil {
		id, err := s.historyService.LogEmailAttempt(reminderID, userID, to, subject, body)
		if err != nil {
			s.logger.Warn("Failed to log email attempt", map[string]interface{}{
				"user_id":     userID,
				"reminder_id": reminderID,
				"error":       err.Error(),
			})
			// Continue with send even if logging fails
		} else {
			logID = id
		}
	}

	if settings == nil {
		if logID > 0 && s.historyService != nil {
			if logErr := s.historyService.LogEmailFailure(logID, "SMTP settings cannot be nil"); logErr != nil {
				s.logger.Warn("Failed to log email failure", map[string]interface{}{
					"log_id": logID,
					"error":  logErr.Error(),
				})
			}
		}
		return fmt.Errorf("SMTP settings cannot be nil")
	}

	// Validate SMTP settings
	if err := s.ValidateSMTPSettings(settings); err != nil {
		if logID > 0 && s.historyService != nil {
			if logErr := s.historyService.LogEmailFailure(logID, fmt.Sprintf("invalid SMTP settings: %v", err)); logErr != nil {
				s.logger.Warn("Failed to log email failure", map[string]interface{}{
					"log_id": logID,
					"error":  logErr.Error(),
				})
			}
		}
		return fmt.Errorf("invalid SMTP settings: %w", err)
	}

	// Validate recipients
	for _, email := range to {
		if !s.ValidateEmailAddress(email) {
			if logID > 0 && s.historyService != nil {
				if logErr := s.historyService.LogEmailFailure(logID, fmt.Sprintf("invalid email address: %s", email)); logErr != nil {
					s.logger.Warn("Failed to log email failure", map[string]interface{}{
						"log_id": logID,
						"error":  logErr.Error(),
					})
				}
			}
			return fmt.Errorf("invalid email address: %s", email)
		}
	}

	// Format email message with proper headers
	from := settings.FromEmail
	if settings.FromName != "" {
		from = fmt.Sprintf("%s <%s>", settings.FromName, settings.FromEmail)
	}

	message := s.formatEmailMessage(from, to, subject, body)

	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%d", settings.Host, settings.Port)

	// Establish connection
	var client *smtp.Client
	var err error

	// Port 465 uses implicit TLS (direct TLS connection)
	// Port 587 uses STARTTLS (plain connection then upgrade to TLS)
	if settings.Port == 465 {
		// Use direct TLS connection for port 465
		tlsConfig := &tls.Config{
			ServerName: settings.Host,
		}

		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			s.logger.Error("Failed to establish TLS connection", map[string]interface{}{
				"host":  settings.Host,
				"port":  settings.Port,
				"error": err.Error(),
			})
			if logID > 0 && s.historyService != nil {
				if logErr := s.historyService.LogEmailFailure(logID, fmt.Sprintf("failed to establish TLS connection: %v", err)); logErr != nil {
					s.logger.Warn("Failed to log email failure", map[string]interface{}{
						"log_id": logID,
						"error":  logErr.Error(),
					})
				}
			}
			return fmt.Errorf("failed to establish TLS connection: %w", err)
		}
		defer conn.Close()

		client, err = smtp.NewClient(conn, settings.Host)
		if err != nil {
			s.logger.Error("Failed to create SMTP client", map[string]interface{}{
				"host":  settings.Host,
				"error": err.Error(),
			})
			if logID > 0 && s.historyService != nil {
				if logErr := s.historyService.LogEmailFailure(logID, fmt.Sprintf("failed to create SMTP client: %v", err)); logErr != nil {
					s.logger.Warn("Failed to log email failure", map[string]interface{}{
						"log_id": logID,
						"error":  logErr.Error(),
					})
				}
			}
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
	} else {
		// Use plain connection with STARTTLS for other ports (typically 587 or 25)
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			s.logger.Error("Failed to connect to SMTP server", map[string]interface{}{
				"host":  settings.Host,
				"port":  settings.Port,
				"error": err.Error(),
			})
			if logID > 0 && s.historyService != nil {
				if logErr := s.historyService.LogEmailFailure(logID, fmt.Sprintf("failed to connect to SMTP server: %v", err)); logErr != nil {
					s.logger.Warn("Failed to log email failure", map[string]interface{}{
						"log_id": logID,
						"error":  logErr.Error(),
					})
				}
			}
			return fmt.Errorf("failed to connect to SMTP server: %w", err)
		}
		defer conn.Close()

		client, err = smtp.NewClient(conn, settings.Host)
		if err != nil {
			s.logger.Error("Failed to create SMTP client", map[string]interface{}{
				"host":  settings.Host,
				"error": err.Error(),
			})
			if logID > 0 && s.historyService != nil {
				if logErr := s.historyService.LogEmailFailure(logID, fmt.Sprintf("failed to create SMTP client: %v", err)); logErr != nil {
					s.logger.Warn("Failed to log email failure", map[string]interface{}{
						"log_id": logID,
						"error":  logErr.Error(),
					})
				}
			}
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}

		// Use STARTTLS if enabled and supported
		if settings.UseTLS {
			if ok, _ := client.Extension("STARTTLS"); ok {
				tlsConfig := &tls.Config{
					ServerName: settings.Host,
				}
				if err := client.StartTLS(tlsConfig); err != nil {
					s.logger.Warn("Failed to start TLS", map[string]interface{}{
						"host":  settings.Host,
						"error": err.Error(),
					})
					if logID > 0 && s.historyService != nil {
						if logErr := s.historyService.LogEmailFailure(logID, fmt.Sprintf("failed to start TLS: %v", err)); logErr != nil {
							s.logger.Warn("Failed to log email failure", map[string]interface{}{
								"log_id": logID,
								"error":  logErr.Error(),
							})
						}
					}
					return fmt.Errorf("failed to start TLS: %w", err)
				}
			} else {
				// STARTTLS not supported but UseTLS is enabled
				s.logger.Error("STARTTLS not supported by server", map[string]interface{}{
					"host": settings.Host,
				})
				if logID > 0 && s.historyService != nil {
					if logErr := s.historyService.LogEmailFailure(logID, "STARTTLS not supported by server but TLS is required"); logErr != nil {
						s.logger.Warn("Failed to log email failure", map[string]interface{}{
							"log_id": logID,
							"error":  logErr.Error(),
						})
					}
				}
				return fmt.Errorf("STARTTLS not supported by server but TLS is required")
			}
		}
	}
	defer client.Close()

	// Authenticate with SMTP server
	if settings.Username != "" && settings.Password != "" {
		auth := smtp.PlainAuth("", settings.Username, settings.Password, settings.Host)
		if err := client.Auth(auth); err != nil {
			s.logger.Error("SMTP authentication failed", map[string]interface{}{
				"host":     settings.Host,
				"username": settings.Username,
				"error":    err.Error(),
			})
			if logID > 0 && s.historyService != nil {
				if logErr := s.historyService.LogEmailFailure(logID, fmt.Sprintf("SMTP authentication failed: %v", err)); logErr != nil {
					s.logger.Warn("Failed to log email failure", map[string]interface{}{
						"log_id": logID,
						"error":  logErr.Error(),
					})
				}
			}
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
	}

	// Set sender
	if err := client.Mail(settings.FromEmail); err != nil {
		if logID > 0 && s.historyService != nil {
			if logErr := s.historyService.LogEmailFailure(logID, fmt.Sprintf("failed to set sender: %v", err)); logErr != nil {
				s.logger.Warn("Failed to log email failure", map[string]interface{}{
					"log_id": logID,
					"error":  logErr.Error(),
				})
			}
		}
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipients
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			if logID > 0 && s.historyService != nil {
				if logErr := s.historyService.LogEmailFailure(logID, fmt.Sprintf("failed to set recipient %s: %v", recipient, err)); logErr != nil {
					s.logger.Warn("Failed to log email failure", map[string]interface{}{
						"log_id": logID,
						"error":  logErr.Error(),
					})
				}
			}
			return fmt.Errorf("failed to set recipient %s: %w", recipient, err)
		}
	}

	// Send message body
	writer, err := client.Data()
	if err != nil {
		if logID > 0 && s.historyService != nil {
			if logErr := s.historyService.LogEmailFailure(logID, fmt.Sprintf("failed to open data writer: %v", err)); logErr != nil {
				s.logger.Warn("Failed to log email failure", map[string]interface{}{
					"log_id": logID,
					"error":  logErr.Error(),
				})
			}
		}
		return fmt.Errorf("failed to open data writer: %w", err)
	}

	_, err = writer.Write([]byte(message))
	if err != nil {
		writer.Close()
		if logID > 0 && s.historyService != nil {
			if logErr := s.historyService.LogEmailFailure(logID, fmt.Sprintf("failed to write message: %v", err)); logErr != nil {
				s.logger.Warn("Failed to log email failure", map[string]interface{}{
					"log_id": logID,
					"error":  logErr.Error(),
				})
			}
		}
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = writer.Close()
	if err != nil {
		if logID > 0 && s.historyService != nil {
			if logErr := s.historyService.LogEmailFailure(logID, fmt.Sprintf("failed to close data writer: %v", err)); logErr != nil {
				s.logger.Warn("Failed to log email failure", map[string]interface{}{
					"log_id": logID,
					"error":  logErr.Error(),
				})
			}
		}
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	// Send QUIT command
	if err := client.Quit(); err != nil {
		if logID > 0 && s.historyService != nil {
			if logErr := s.historyService.LogEmailFailure(logID, fmt.Sprintf("failed to quit SMTP session: %v", err)); logErr != nil {
				s.logger.Warn("Failed to log email failure", map[string]interface{}{
					"log_id": logID,
					"error":  logErr.Error(),
				})
			}
		}
		return fmt.Errorf("failed to quit SMTP session: %w", err)
	}

	// Log success after successful send
	if logID > 0 && s.historyService != nil {
		if err := s.historyService.LogEmailSuccess(logID); err != nil {
			s.logger.Warn("Failed to log email success", map[string]interface{}{
				"log_id": logID,
				"error":  err.Error(),
			})
			// Don't fail the send operation if logging fails
		}
	}

	return nil
}

// formatEmailMessage formats the email with proper headers
func (s *EmailService) formatEmailMessage(from string, to []string, subject, body string) string {
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = strings.Join(to, ", ")
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/plain; charset=UTF-8"

	var message strings.Builder
	for key, value := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	message.WriteString("\r\n")
	message.WriteString(body)

	return message.String()
}

// ValidateEmailAddress validates an email address format using RFC 5322 regex
func (s *EmailService) ValidateEmailAddress(email string) bool {
	// RFC 5322 compliant email regex (simplified but practical)
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// ValidateSMTPSettings validates SMTP configuration parameters
func (s *EmailService) ValidateSMTPSettings(settings *models.SMTPSettings) error {
	if settings.Host == "" {
		return fmt.Errorf("SMTP host is required")
	}

	if settings.Port < 1 || settings.Port > 65535 {
		return fmt.Errorf("SMTP port must be between 1 and 65535")
	}

	if settings.Username == "" {
		return fmt.Errorf("SMTP username is required")
	}

	if settings.Password == "" {
		return fmt.Errorf("SMTP password is required")
	}

	if settings.FromEmail == "" {
		return fmt.Errorf("from email is required")
	}

	if !s.ValidateEmailAddress(settings.FromEmail) {
		return fmt.Errorf("from email has invalid format")
	}

	return nil
}

// TestSMTPConnection tests SMTP connection and authentication
func (s *EmailService) TestSMTPConnection(settings *models.SMTPSettings) error {
	// Validate settings first
	if err := s.ValidateSMTPSettings(settings); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	addr := fmt.Sprintf("%s:%d", settings.Host, settings.Port)
	
	s.logger.Info("Testing SMTP connection", map[string]interface{}{
		"host": settings.Host,
		"port": settings.Port,
		"username": settings.Username,
	})

	// Establish connection
	var client *smtp.Client

	// Port 465 uses implicit TLS (direct TLS connection)
	// Port 587 uses STARTTLS (plain connection then upgrade to TLS)
	if settings.Port == 465 {
		// Use direct TLS connection for port 465
		tlsConfig := &tls.Config{
			ServerName: settings.Host,
		}

		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("TLS connection failed: %w", err)
		}
		defer conn.Close()

		client, err = smtp.NewClient(conn, settings.Host)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
	} else {
		// Use plain connection with STARTTLS for other ports
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			return fmt.Errorf("connection failed: %w", err)
		}
		defer conn.Close()

		client, err = smtp.NewClient(conn, settings.Host)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}

		// Use STARTTLS if enabled and supported
		if settings.UseTLS {
			if ok, _ := client.Extension("STARTTLS"); ok {
				tlsConfig := &tls.Config{
					ServerName: settings.Host,
				}
				if err := client.StartTLS(tlsConfig); err != nil {
					return fmt.Errorf("STARTTLS failed: %w", err)
				}
			} else {
				// STARTTLS not supported but UseTLS is enabled
				return fmt.Errorf("STARTTLS not supported by server but TLS is required")
			}
		}
	}
	defer client.Close()

	// Test authentication
	if settings.Username != "" && settings.Password != "" {
		auth := smtp.PlainAuth("", settings.Username, settings.Password, settings.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}
	}

	// Send QUIT command
	if err := client.Quit(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}

	s.logger.Info("SMTP connection test successful", map[string]interface{}{
		"host": settings.Host,
		"port": settings.Port,
	})

	return nil
}
