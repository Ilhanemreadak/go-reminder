package services

import (
	"database/sql"
	"testing"
	"time"

	"email-reminder-system/database"
	"email-reminder-system/models"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) (*sql.DB, database.Repository) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	// Create schema
	schema := `
	CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE sessions (
		id TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL,
		expires_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	`

	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	repo := database.NewSQLiteRepository(db)
	return db, repo
}

func TestHashPassword(t *testing.T) {
	_, repo := setupTestDB(t)
	authService := NewAuthService(repo, 24*time.Hour)

	password := "testpassword123"
	hash, err := authService.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Fatal("HashPassword returned empty hash")
	}

	if hash == password {
		t.Fatal("HashPassword returned plaintext password")
	}
}

func TestVerifyPassword(t *testing.T) {
	_, repo := setupTestDB(t)
	authService := NewAuthService(repo, 24*time.Hour)

	password := "testpassword123"
	hash, err := authService.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Test correct password
	if !authService.VerifyPassword(hash, password) {
		t.Fatal("VerifyPassword failed for correct password")
	}

	// Test incorrect password
	if authService.VerifyPassword(hash, "wrongpassword") {
		t.Fatal("VerifyPassword succeeded for incorrect password")
	}
}

func TestLogin_Success(t *testing.T) {
	_, repo := setupTestDB(t)
	authService := NewAuthService(repo, 24*time.Hour)

	// Create a test user
	hash, _ := authService.HashPassword("testpass")
	user := &models.User{
		Username:     "testuser",
		PasswordHash: hash,
	}
	err := repo.CreateUser(user)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Test login
	session, err := authService.Login("testuser", "testpass")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if session == nil {
		t.Fatal("Login returned nil session")
	}

	if session.ID == "" {
		t.Fatal("Session ID is empty")
	}

	if session.UserID != user.ID {
		t.Fatalf("Session UserID mismatch: got %d, want %d", session.UserID, user.ID)
	}

	if session.ExpiresAt.Before(time.Now()) {
		t.Fatal("Session already expired")
	}
}

func TestLogin_InvalidUsername(t *testing.T) {
	_, repo := setupTestDB(t)
	authService := NewAuthService(repo, 24*time.Hour)

	_, err := authService.Login("nonexistent", "password")
	if err == nil {
		t.Fatal("Login should fail for invalid username")
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	_, repo := setupTestDB(t)
	authService := NewAuthService(repo, 24*time.Hour)

	// Create a test user
	hash, _ := authService.HashPassword("correctpass")
	user := &models.User{
		Username:     "testuser",
		PasswordHash: hash,
	}
	repo.CreateUser(user)

	// Test login with wrong password
	_, err := authService.Login("testuser", "wrongpass")
	if err == nil {
		t.Fatal("Login should fail for invalid password")
	}
}

func TestValidateSession_Valid(t *testing.T) {
	_, repo := setupTestDB(t)
	authService := NewAuthService(repo, 24*time.Hour)

	// Create a test user
	hash, _ := authService.HashPassword("testpass")
	user := &models.User{
		Username:     "testuser",
		PasswordHash: hash,
	}
	repo.CreateUser(user)

	// Login to create session
	session, _ := authService.Login("testuser", "testpass")

	// Validate session
	validatedUser, err := authService.ValidateSession(session.ID)
	if err != nil {
		t.Fatalf("ValidateSession failed: %v", err)
	}

	if validatedUser.ID != user.ID {
		t.Fatalf("ValidateSession returned wrong user: got %d, want %d", validatedUser.ID, user.ID)
	}
}

func TestValidateSession_Expired(t *testing.T) {
	_, repo := setupTestDB(t)
	authService := NewAuthService(repo, -1*time.Hour) // Negative duration = already expired

	// Create a test user
	hash, _ := authService.HashPassword("testpass")
	user := &models.User{
		Username:     "testuser",
		PasswordHash: hash,
	}
	repo.CreateUser(user)

	// Login to create expired session
	session, _ := authService.Login("testuser", "testpass")

	// Validate session should fail
	_, err := authService.ValidateSession(session.ID)
	if err == nil {
		t.Fatal("ValidateSession should fail for expired session")
	}
}

func TestValidateSession_Invalid(t *testing.T) {
	_, repo := setupTestDB(t)
	authService := NewAuthService(repo, 24*time.Hour)

	// Validate non-existent session
	_, err := authService.ValidateSession("invalid-session-id")
	if err == nil {
		t.Fatal("ValidateSession should fail for invalid session ID")
	}
}

func TestLogout(t *testing.T) {
	_, repo := setupTestDB(t)
	authService := NewAuthService(repo, 24*time.Hour)

	// Create a test user and login
	hash, _ := authService.HashPassword("testpass")
	user := &models.User{
		Username:     "testuser",
		PasswordHash: hash,
	}
	repo.CreateUser(user)
	session, _ := authService.Login("testuser", "testpass")

	// Logout
	err := authService.Logout(session.ID)
	if err != nil {
		t.Fatalf("Logout failed: %v", err)
	}

	// Validate session should fail after logout
	_, err = authService.ValidateSession(session.ID)
	if err == nil {
		t.Fatal("ValidateSession should fail after logout")
	}
}
