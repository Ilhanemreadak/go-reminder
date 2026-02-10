package services

import (
	"testing"
	"time"

	"email-reminder-system/models"
)

// Mock repository for testing
type mockReminderRepo struct {
	reminders map[int64]*models.Reminder
	nextID    int64
}

func newMockReminderRepo() *mockReminderRepo {
	return &mockReminderRepo{
		reminders: make(map[int64]*models.Reminder),
		nextID:    1,
	}
}

func (m *mockReminderRepo) CreateReminder(reminder *models.Reminder) error {
	reminder.ID = m.nextID
	m.nextID++
	reminder.CreatedAt = time.Now()
	reminder.UpdatedAt = time.Now()
	m.reminders[reminder.ID] = reminder
	return nil
}

func (m *mockReminderRepo) GetReminder(id int64) (*models.Reminder, error) {
	if r, ok := m.reminders[id]; ok {
		return r, nil
	}
	return nil, nil
}

func (m *mockReminderRepo) GetRemindersByUserID(userID int64) ([]*models.Reminder, error) {
	var result []*models.Reminder
	for _, r := range m.reminders {
		if r.UserID == userID {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *mockReminderRepo) GetDueReminders(now time.Time) ([]*models.Reminder, error) {
	return nil, nil
}

func (m *mockReminderRepo) UpdateReminder(reminder *models.Reminder) error {
	if _, ok := m.reminders[reminder.ID]; ok {
		reminder.UpdatedAt = time.Now()
		m.reminders[reminder.ID] = reminder
		return nil
	}
	return nil
}

func (m *mockReminderRepo) DeleteReminder(id int64) error {
	delete(m.reminders, id)
	return nil
}

func (m *mockReminderRepo) GetUserByUsername(username string) (*models.User, error) { return nil, nil }
func (m *mockReminderRepo) GetUserByID(id int64) (*models.User, error)              { return nil, nil }
func (m *mockReminderRepo) CreateUser(user *models.User) error                      { return nil }
func (m *mockReminderRepo) CreateSession(session *models.Session) error             { return nil }
func (m *mockReminderRepo) GetSession(id string) (*models.Session, error)           { return nil, nil }
func (m *mockReminderRepo) DeleteSession(id string) error                           { return nil }
func (m *mockReminderRepo) CleanExpiredSessions() error                             { return nil }
func (m *mockReminderRepo) GetSMTPSettings(userID int64) (*models.SMTPSettings, error) {
	return nil, nil
}
func (m *mockReminderRepo) UpsertSMTPSettings(settings *models.SMTPSettings) error { return nil }

func TestValidateReminder_RequiredFields(t *testing.T) {
	service := NewReminderService(newMockReminderRepo())

	tests := []struct {
		name        string
		reminder    *models.Reminder
		expectError bool
	}{
		{
			name: "valid daily reminder",
			reminder: &models.Reminder{
				Title:        "Test Reminder",
				Recipients:   []string{"test@example.com"},
				EmailContent: "Test content",
				ScheduleType: "daily",
				TimeOfDay:    "09:00",
			},
			expectError: false,
		},
		{
			name: "missing title",
			reminder: &models.Reminder{
				Recipients:   []string{"test@example.com"},
				EmailContent: "Test content",
				ScheduleType: "daily",
				TimeOfDay:    "09:00",
			},
			expectError: true,
		},
		{
			name: "missing recipients",
			reminder: &models.Reminder{
				Title:        "Test Reminder",
				Recipients:   []string{},
				EmailContent: "Test content",
				ScheduleType: "daily",
				TimeOfDay:    "09:00",
			},
			expectError: true,
		},
		{
			name: "missing email content",
			reminder: &models.Reminder{
				Title:        "Test Reminder",
				Recipients:   []string{"test@example.com"},
				ScheduleType: "daily",
				TimeOfDay:    "09:00",
			},
			expectError: true,
		},
		{
			name: "invalid schedule type",
			reminder: &models.Reminder{
				Title:        "Test Reminder",
				Recipients:   []string{"test@example.com"},
				EmailContent: "Test content",
				ScheduleType: "invalid",
				TimeOfDay:    "09:00",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateReminder(tt.reminder)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateReminder_EmailValidation(t *testing.T) {
	service := NewReminderService(newMockReminderRepo())

	tests := []struct {
		name        string
		email       string
		expectError bool
	}{
		{"valid email", "test@example.com", false},
		{"valid email with subdomain", "user@mail.example.com", false},
		{"valid email with plus", "user+tag@example.com", false},
		{"invalid email no @", "testexample.com", true},
		{"invalid email no domain", "test@", true},
		{"invalid email no local", "@example.com", true},
		{"empty email", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reminder := &models.Reminder{
				Title:        "Test",
				Recipients:   []string{tt.email},
				EmailContent: "Content",
				ScheduleType: "daily",
				TimeOfDay:    "09:00",
			}
			err := service.ValidateReminder(reminder)
			if tt.expectError && err == nil {
				t.Errorf("expected error for email %s but got none", tt.email)
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error for email %s: %v", tt.email, err)
			}
		})
	}
}

func TestValidateReminder_TimeFormat(t *testing.T) {
	service := NewReminderService(newMockReminderRepo())

	tests := []struct {
		name        string
		timeOfDay   string
		expectError bool
	}{
		{"valid time 00:00", "00:00", false},
		{"valid time 09:30", "09:30", false},
		{"valid time 23:59", "23:59", false},
		{"invalid time 24:00", "24:00", true},
		{"invalid time 12:60", "12:60", true},
		{"invalid time format", "9:30", true},
		{"invalid time format 2", "09:5", true},
		{"invalid time no colon", "0930", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reminder := &models.Reminder{
				Title:        "Test",
				Recipients:   []string{"test@example.com"},
				EmailContent: "Content",
				ScheduleType: "daily",
				TimeOfDay:    tt.timeOfDay,
			}
			err := service.ValidateReminder(reminder)
			if tt.expectError && err == nil {
				t.Errorf("expected error for time %s but got none", tt.timeOfDay)
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error for time %s: %v", tt.timeOfDay, err)
			}
		})
	}
}

func TestValidateReminder_ScheduleSpecific(t *testing.T) {
	service := NewReminderService(newMockReminderRepo())

	tests := []struct {
		name        string
		reminder    *models.Reminder
		expectError bool
	}{
		{
			name: "valid weekly reminder",
			reminder: &models.Reminder{
				Title:        "Test",
				Recipients:   []string{"test@example.com"},
				EmailContent: "Content",
				ScheduleType: "weekly",
				DayOfWeek:    3,
				TimeOfDay:    "09:00",
			},
			expectError: false,
		},
		{
			name: "invalid weekly day of week",
			reminder: &models.Reminder{
				Title:        "Test",
				Recipients:   []string{"test@example.com"},
				EmailContent: "Content",
				ScheduleType: "weekly",
				DayOfWeek:    7,
				TimeOfDay:    "09:00",
			},
			expectError: true,
		},
		{
			name: "valid monthly reminder",
			reminder: &models.Reminder{
				Title:        "Test",
				Recipients:   []string{"test@example.com"},
				EmailContent: "Content",
				ScheduleType: "monthly",
				DayOfMonth:   15,
				TimeOfDay:    "09:00",
			},
			expectError: false,
		},
		{
			name: "invalid monthly day of month",
			reminder: &models.Reminder{
				Title:        "Test",
				Recipients:   []string{"test@example.com"},
				EmailContent: "Content",
				ScheduleType: "monthly",
				DayOfMonth:   32,
				TimeOfDay:    "09:00",
			},
			expectError: true,
		},
		{
			name: "valid custom reminder",
			reminder: &models.Reminder{
				Title:        "Test",
				Recipients:   []string{"test@example.com"},
				EmailContent: "Content",
				ScheduleType: "custom",
				IntervalDays: 5,
				TimeOfDay:    "09:00",
			},
			expectError: false,
		},
		{
			name: "invalid custom interval",
			reminder: &models.Reminder{
				Title:        "Test",
				Recipients:   []string{"test@example.com"},
				EmailContent: "Content",
				ScheduleType: "custom",
				IntervalDays: 0,
				TimeOfDay:    "09:00",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateReminder(tt.reminder)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestCalculateNextSendTime_Daily(t *testing.T) {
	service := NewReminderService(newMockReminderRepo())

	now := time.Now()
	reminder := &models.Reminder{
		ScheduleType: "daily",
		TimeOfDay:    "14:30",
	}

	nextSend, err := service.CalculateNextSendTime(reminder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if nextSend.Hour() != 14 || nextSend.Minute() != 30 {
		t.Errorf("expected time 14:30, got %02d:%02d", nextSend.Hour(), nextSend.Minute())
	}

	if nextSend.Before(now) {
		t.Errorf("next send time should be in the future")
	}
}

func TestCalculateNextSendTime_Weekly(t *testing.T) {
	service := NewReminderService(newMockReminderRepo())

	now := time.Now()
	reminder := &models.Reminder{
		ScheduleType: "weekly",
		DayOfWeek:    1, // Monday
		TimeOfDay:    "10:00",
	}

	nextSend, err := service.CalculateNextSendTime(reminder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if nextSend.Weekday() != time.Monday {
		t.Errorf("expected Monday, got %v", nextSend.Weekday())
	}

	if nextSend.Hour() != 10 || nextSend.Minute() != 0 {
		t.Errorf("expected time 10:00, got %02d:%02d", nextSend.Hour(), nextSend.Minute())
	}

	if nextSend.Before(now) {
		t.Errorf("next send time should be in the future")
	}
}

func TestCalculateNextSendTime_Monthly(t *testing.T) {
	service := NewReminderService(newMockReminderRepo())

	now := time.Now()
	reminder := &models.Reminder{
		ScheduleType: "monthly",
		DayOfMonth:   15,
		TimeOfDay:    "12:00",
	}

	nextSend, err := service.CalculateNextSendTime(reminder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if nextSend.Day() != 15 {
		t.Errorf("expected day 15, got %d", nextSend.Day())
	}

	if nextSend.Hour() != 12 || nextSend.Minute() != 0 {
		t.Errorf("expected time 12:00, got %02d:%02d", nextSend.Hour(), nextSend.Minute())
	}

	if nextSend.Before(now) {
		t.Errorf("next send time should be in the future")
	}
}

func TestCalculateNextSendTime_Custom(t *testing.T) {
	service := NewReminderService(newMockReminderRepo())

	now := time.Now()
	reminder := &models.Reminder{
		ScheduleType: "custom",
		IntervalDays: 7,
		TimeOfDay:    "08:00",
	}

	nextSend, err := service.CalculateNextSendTime(reminder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if nextSend.Hour() != 8 || nextSend.Minute() != 0 {
		t.Errorf("expected time 08:00, got %02d:%02d", nextSend.Hour(), nextSend.Minute())
	}

	if nextSend.Before(now) {
		t.Errorf("next send time should be in the future")
	}
}

func TestCreateReminder(t *testing.T) {
	repo := newMockReminderRepo()
	service := NewReminderService(repo)

	reminder := &models.Reminder{
		Title:        "Test Reminder",
		Recipients:   []string{"test@example.com"},
		EmailContent: "Test content",
		ScheduleType: "daily",
		TimeOfDay:    "09:00",
	}

	err := service.CreateReminder(1, reminder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if reminder.ID == 0 {
		t.Error("reminder ID should be set")
	}

	if reminder.UserID != 1 {
		t.Errorf("expected user ID 1, got %d", reminder.UserID)
	}

	if !reminder.IsActive {
		t.Error("reminder should be active")
	}

	if reminder.NextSendAt.IsZero() {
		t.Error("next send time should be calculated")
	}
}

func TestGetReminder_Authorization(t *testing.T) {
	repo := newMockReminderRepo()
	service := NewReminderService(repo)

	// Create reminder for user 1
	reminder := &models.Reminder{
		Title:        "Test",
		Recipients:   []string{"test@example.com"},
		EmailContent: "Content",
		ScheduleType: "daily",
		TimeOfDay:    "09:00",
	}
	service.CreateReminder(1, reminder)

	// Try to get as user 1 (should succeed)
	retrieved, err := service.GetReminder(reminder.ID, 1)
	if err != nil {
		t.Errorf("user 1 should be able to access their reminder: %v", err)
	}
	if retrieved == nil {
		t.Error("expected reminder to be returned")
	}

	// Try to get as user 2 (should fail)
	_, err = service.GetReminder(reminder.ID, 2)
	if err == nil {
		t.Error("user 2 should not be able to access user 1's reminder")
	}
}

func TestUpdateReminder(t *testing.T) {
	repo := newMockReminderRepo()
	service := NewReminderService(repo)

	// Create reminder
	reminder := &models.Reminder{
		Title:        "Original Title",
		Recipients:   []string{"test@example.com"},
		EmailContent: "Original content",
		ScheduleType: "daily",
		TimeOfDay:    "09:00",
	}
	service.CreateReminder(1, reminder)

	// Update reminder
	reminder.Title = "Updated Title"
	err := service.UpdateReminder(reminder, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify update
	retrieved, _ := service.GetReminder(reminder.ID, 1)
	if retrieved.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got '%s'", retrieved.Title)
	}
}

func TestDeleteReminder(t *testing.T) {
	repo := newMockReminderRepo()
	service := NewReminderService(repo)

	// Create reminder
	reminder := &models.Reminder{
		Title:        "Test",
		Recipients:   []string{"test@example.com"},
		EmailContent: "Content",
		ScheduleType: "daily",
		TimeOfDay:    "09:00",
	}
	service.CreateReminder(1, reminder)

	// Delete reminder
	err := service.DeleteReminder(reminder.ID, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify deletion
	retrieved, _ := repo.GetReminder(reminder.ID)
	if retrieved != nil {
		t.Error("reminder should be deleted")
	}
}
