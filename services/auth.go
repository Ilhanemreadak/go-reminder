package services

import (
	"fmt"
	"time"

	"email-reminder-system/database"
	"email-reminder-system/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	Login(username, password string) (*models.Session, error)
	Logout(sessionID string) error
	ValidateSession(sessionID string) (*models.User, error)
	HashPassword(password string) (string, error)
	VerifyPassword(hash, password string) bool
}

// authService implements AuthService interface
type authService struct {
	repo              database.Repository
	sessionDuration   time.Duration
	logger            *Logger
}

// NewAuthService creates a new authentication service
func NewAuthService(repo database.Repository, sessionDuration time.Duration) AuthService {
	return &authService{
		repo:            repo,
		sessionDuration: sessionDuration,
		logger:          GetLogger(),
	}
}

// Login authenticates a user with username and password
// Returns a new session if credentials are valid
func (s *authService) Login(username, password string) (*models.Session, error) {
	// Get user by username
	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		s.logger.LogAuthAttempt(username, false, fmt.Errorf("invalid credentials"))
		return nil, fmt.Errorf("invalid credentials")
	}

	// Verify password
	if !s.VerifyPassword(user.PasswordHash, password) {
		s.logger.LogAuthAttempt(username, false, fmt.Errorf("invalid credentials"))
		return nil, fmt.Errorf("invalid credentials")
	}

	// Create new session
	session := &models.Session{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(s.sessionDuration),
	}

	err = s.repo.CreateSession(session)
	if err != nil {
		s.logger.LogDatabaseError("CreateSession", err, map[string]interface{}{
			"user_id": user.ID,
		})
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	s.logger.LogAuthAttempt(username, true, nil)
	return session, nil
}

// Logout invalidates a session
func (s *authService) Logout(sessionID string) error {
	err := s.repo.DeleteSession(sessionID)
	if err != nil {
		s.logger.LogDatabaseError("DeleteSession", err, map[string]interface{}{
			"session_id": sessionID,
		})
		return fmt.Errorf("failed to logout: %w", err)
	}
	s.logger.Info("User logged out", map[string]interface{}{
		"session_id": sessionID,
	})
	return nil
}

// ValidateSession checks if a session is valid and not expired
// Returns the associated user if valid
func (s *authService) ValidateSession(sessionID string) (*models.User, error) {
	// Get session from repository
	session, err := s.repo.GetSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session")
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		// Clean up expired session
		s.repo.DeleteSession(sessionID)
		return nil, fmt.Errorf("session expired")
	}

	// Get user associated with session
	user, err := s.repo.GetUserByID(session.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// HashPassword hashes a password using bcrypt with cost 12
func (s *authService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword verifies a password against a bcrypt hash
func (s *authService) VerifyPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
