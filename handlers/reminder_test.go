package handlers

import (
	"database/sql"
	"email-reminder-system/database"
	"email-reminder-system/models"
	"email-reminder-system/services"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// setupTestDB creates an in-memory database for testing
func setupTestDB(t *testing.T) (*sql.DB, database.Repository) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	// Execute schema
	if _, err := db.Exec(database.GetSchema()); err != nil {
		t.Fatalf("Failed to execute schema: %v", err)
	}

	repo := database.NewSQLiteRepository(db)
	return db, repo
}

// createTestUser creates a test user and returns the user
func createTestUser(t *testing.T, repo database.Repository) *models.User {
	authService := services.NewAuthService(repo, 24*time.Hour)
	hash, err := authService.HashPassword("testpass")
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	user := &models.User{
		Username:     "testuser",
		PasswordHash: hash,
	}

	err = repo.CreateUser(user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return user
}

// createTestSession creates a test session for a user
func createTestSession(t *testing.T, repo database.Repository, userID int64) *models.Session {
	session := &models.Session{
		ID:        "test-session-id",
		UserID:    userID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	err := repo.CreateSession(session)
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	return session
}

func TestListReminders(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	user := createTestUser(t, repo)
	session := createTestSession(t, repo, user.ID)

	// Create test reminder
	reminder := &models.Reminder{
		UserID:       user.ID,
		Title:        "Test Reminder",
		Recipients:   []string{"test@example.com"},
		EmailContent: "Test content",
		ScheduleType: "daily",
		TimeOfDay:    "14:00",
		NextSendAt:   time.Now().Add(24 * time.Hour),
		IsActive:     true,
	}
	err := repo.CreateReminder(reminder)
	if err != nil {
		t.Fatalf("Failed to create test reminder: %v", err)
	}

	// Create handler
	reminderService := services.NewReminderService(repo)
	templates := template.Must(template.ParseGlob("../templates/*.html"))
	handler := NewReminderHandler(reminderService, templates)

	// Create request with session cookie
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: session.ID,
	})

	// Add user to context (simulating middleware)
	authService := services.NewAuthService(repo, 24*time.Hour)
	middleware := NewAuthMiddleware(authService)
	
	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler through middleware
	middleware.RequireAuth(handler.ListReminders).ServeHTTP(rr, req)

	// Check response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check that response contains reminder title
	body := rr.Body.String()
	if !strings.Contains(body, "Test Reminder") {
		t.Errorf("Response does not contain reminder title")
	}
}

func TestCreateReminder(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	user := createTestUser(t, repo)
	session := createTestSession(t, repo, user.ID)

	// Create handler
	reminderService := services.NewReminderService(repo)
	templates := template.Must(template.ParseGlob("../templates/*.html"))
	handler := NewReminderHandler(reminderService, templates)

	// Create form data
	formData := url.Values{}
	formData.Set("title", "New Reminder")
	formData.Set("recipients", "test1@example.com, test2@example.com")
	formData.Set("email_content", "Test email content")
	formData.Set("schedule_type", "daily")
	formData.Set("time_of_day", "15:30")

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/reminders", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: session.ID,
	})

	// Add user to context
	authService := services.NewAuthService(repo, 24*time.Hour)
	middleware := NewAuthMiddleware(authService)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler through middleware
	middleware.RequireAuth(handler.CreateReminder).ServeHTTP(rr, req)

	// Check response - should redirect to list page
	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}

	// Verify reminder was created
	reminders, err := repo.GetRemindersByUserID(user.ID)
	if err != nil {
		t.Fatalf("Failed to get reminders: %v", err)
	}

	if len(reminders) != 1 {
		t.Errorf("Expected 1 reminder, got %d", len(reminders))
	}

	if reminders[0].Title != "New Reminder" {
		t.Errorf("Expected title 'New Reminder', got '%s'", reminders[0].Title)
	}
}
