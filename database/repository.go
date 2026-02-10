package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"email-reminder-system/models"
)

// Repository defines the interface for data access operations
type Repository interface {
	// User operations
	GetUserByUsername(username string) (*models.User, error)
	GetUserByID(id int64) (*models.User, error)
	CreateUser(user *models.User) error

	// Session operations
	CreateSession(session *models.Session) error
	GetSession(id string) (*models.Session, error)
	DeleteSession(id string) error
	CleanExpiredSessions() error

	// Reminder operations
	CreateReminder(reminder *models.Reminder) error
	GetReminder(id int64) (*models.Reminder, error)
	GetRemindersByUserID(userID int64) ([]*models.Reminder, error)
	GetDueReminders(now time.Time) ([]*models.Reminder, error)
	UpdateReminder(reminder *models.Reminder) error
	DeleteReminder(id int64) error

	// SMTP Settings operations
	GetSMTPSettings(userID int64) (*models.SMTPSettings, error)
	UpsertSMTPSettings(settings *models.SMTPSettings) error
}

// SQLiteRepository implements Repository interface using SQLite
type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository creates a new SQLite repository
func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

// User operations

func (r *SQLiteRepository) GetUserByUsername(username string) (*models.User, error) {
	query := `SELECT id, username, password_hash, created_at FROM users WHERE username = ?`
	user := &models.User{}
	err := r.db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (r *SQLiteRepository) GetUserByID(id int64) (*models.User, error) {
	query := `SELECT id, username, password_hash, created_at FROM users WHERE id = ?`
	user := &models.User{}
	err := r.db.QueryRow(query, id).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (r *SQLiteRepository) CreateUser(user *models.User) error {
	query := `INSERT INTO users (username, password_hash) VALUES (?, ?)`
	result, err := r.db.Exec(query, user.Username, user.PasswordHash)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get user ID: %w", err)
	}
	user.ID = id
	user.CreatedAt = time.Now()
	return nil
}

// Session operations

func (r *SQLiteRepository) CreateSession(session *models.Session) error {
	query := `INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)`
	_, err := r.db.Exec(query, session.ID, session.UserID, session.ExpiresAt)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	session.CreatedAt = time.Now()
	return nil
}

func (r *SQLiteRepository) GetSession(id string) (*models.Session, error) {
	query := `SELECT id, user_id, expires_at, created_at FROM sessions WHERE id = ?`
	session := &models.Session{}
	err := r.db.QueryRow(query, id).Scan(&session.ID, &session.UserID, &session.ExpiresAt, &session.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return session, nil
}

func (r *SQLiteRepository) DeleteSession(id string) error {
	query := `DELETE FROM sessions WHERE id = ?`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) CleanExpiredSessions() error {
	query := `DELETE FROM sessions WHERE expires_at < ?`
	_, err := r.db.Exec(query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to clean expired sessions: %w", err)
	}
	return nil
}

// Reminder operations

func (r *SQLiteRepository) CreateReminder(reminder *models.Reminder) error {
	recipients, err := json.Marshal(reminder.Recipients)
	if err != nil {
		return fmt.Errorf("failed to marshal recipients: %w", err)
	}

	query := `INSERT INTO reminders (user_id, title, recipients, email_content, schedule_type, 
		interval_days, day_of_week, day_of_month, time_of_day, last_sent_at, next_send_at, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query, reminder.UserID, reminder.Title, string(recipients),
		reminder.EmailContent, reminder.ScheduleType, reminder.IntervalDays, reminder.DayOfWeek,
		reminder.DayOfMonth, reminder.TimeOfDay, reminder.LastSentAt, reminder.NextSendAt, reminder.IsActive)
	if err != nil {
		return fmt.Errorf("failed to create reminder: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get reminder ID: %w", err)
	}
	reminder.ID = id
	reminder.CreatedAt = time.Now()
	reminder.UpdatedAt = time.Now()
	return nil
}

func (r *SQLiteRepository) GetReminder(id int64) (*models.Reminder, error) {
	query := `SELECT id, user_id, title, recipients, email_content, schedule_type, interval_days,
		day_of_week, day_of_month, time_of_day, last_sent_at, next_send_at, is_active, created_at, updated_at
		FROM reminders WHERE id = ?`

	reminder := &models.Reminder{}
	var recipientsJSON string
	var lastSentAt sql.NullTime

	err := r.db.QueryRow(query, id).Scan(&reminder.ID, &reminder.UserID, &reminder.Title,
		&recipientsJSON, &reminder.EmailContent, &reminder.ScheduleType, &reminder.IntervalDays,
		&reminder.DayOfWeek, &reminder.DayOfMonth, &reminder.TimeOfDay, &lastSentAt,
		&reminder.NextSendAt, &reminder.IsActive, &reminder.CreatedAt, &reminder.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("reminder not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get reminder: %w", err)
	}

	if err := json.Unmarshal([]byte(recipientsJSON), &reminder.Recipients); err != nil {
		return nil, fmt.Errorf("failed to unmarshal recipients: %w", err)
	}

	if lastSentAt.Valid {
		reminder.LastSentAt = &lastSentAt.Time
	}

	return reminder, nil
}

func (r *SQLiteRepository) GetRemindersByUserID(userID int64) ([]*models.Reminder, error) {
	query := `SELECT id, user_id, title, recipients, email_content, schedule_type, interval_days,
		day_of_week, day_of_month, time_of_day, last_sent_at, next_send_at, is_active, created_at, updated_at
		FROM reminders WHERE user_id = ? ORDER BY created_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reminders: %w", err)
	}
	defer rows.Close()

	var reminders []*models.Reminder
	for rows.Next() {
		reminder := &models.Reminder{}
		var recipientsJSON string
		var lastSentAt sql.NullTime

		err := rows.Scan(&reminder.ID, &reminder.UserID, &reminder.Title, &recipientsJSON,
			&reminder.EmailContent, &reminder.ScheduleType, &reminder.IntervalDays,
			&reminder.DayOfWeek, &reminder.DayOfMonth, &reminder.TimeOfDay, &lastSentAt,
			&reminder.NextSendAt, &reminder.IsActive, &reminder.CreatedAt, &reminder.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reminder: %w", err)
		}

		if err := json.Unmarshal([]byte(recipientsJSON), &reminder.Recipients); err != nil {
			return nil, fmt.Errorf("failed to unmarshal recipients: %w", err)
		}

		if lastSentAt.Valid {
			reminder.LastSentAt = &lastSentAt.Time
		}

		reminders = append(reminders, reminder)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reminders: %w", err)
	}

	return reminders, nil
}

func (r *SQLiteRepository) GetDueReminders(now time.Time) ([]*models.Reminder, error) {
	query := `SELECT id, user_id, title, recipients, email_content, schedule_type, interval_days,
		day_of_week, day_of_month, time_of_day, last_sent_at, next_send_at, is_active, created_at, updated_at
		FROM reminders WHERE next_send_at <= ? AND is_active = 1`

	rows, err := r.db.Query(query, now)
	if err != nil {
		return nil, fmt.Errorf("failed to get due reminders: %w", err)
	}
	defer rows.Close()

	var reminders []*models.Reminder
	for rows.Next() {
		reminder := &models.Reminder{}
		var recipientsJSON string
		var lastSentAt sql.NullTime

		err := rows.Scan(&reminder.ID, &reminder.UserID, &reminder.Title, &recipientsJSON,
			&reminder.EmailContent, &reminder.ScheduleType, &reminder.IntervalDays,
			&reminder.DayOfWeek, &reminder.DayOfMonth, &reminder.TimeOfDay, &lastSentAt,
			&reminder.NextSendAt, &reminder.IsActive, &reminder.CreatedAt, &reminder.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reminder: %w", err)
		}

		if err := json.Unmarshal([]byte(recipientsJSON), &reminder.Recipients); err != nil {
			return nil, fmt.Errorf("failed to unmarshal recipients: %w", err)
		}

		if lastSentAt.Valid {
			reminder.LastSentAt = &lastSentAt.Time
		}

		reminders = append(reminders, reminder)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating due reminders: %w", err)
	}

	return reminders, nil
}

func (r *SQLiteRepository) UpdateReminder(reminder *models.Reminder) error {
	recipients, err := json.Marshal(reminder.Recipients)
	if err != nil {
		return fmt.Errorf("failed to marshal recipients: %w", err)
	}

	query := `UPDATE reminders SET title = ?, recipients = ?, email_content = ?, schedule_type = ?,
		interval_days = ?, day_of_week = ?, day_of_month = ?, time_of_day = ?, last_sent_at = ?,
		next_send_at = ?, is_active = ?, updated_at = ? WHERE id = ?`

	_, err = r.db.Exec(query, reminder.Title, string(recipients), reminder.EmailContent,
		reminder.ScheduleType, reminder.IntervalDays, reminder.DayOfWeek, reminder.DayOfMonth,
		reminder.TimeOfDay, reminder.LastSentAt, reminder.NextSendAt, reminder.IsActive,
		time.Now(), reminder.ID)
	if err != nil {
		return fmt.Errorf("failed to update reminder: %w", err)
	}

	reminder.UpdatedAt = time.Now()
	return nil
}

func (r *SQLiteRepository) DeleteReminder(id int64) error {
	query := `DELETE FROM reminders WHERE id = ?`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete reminder: %w", err)
	}
	return nil
}

// SMTP Settings operations

func (r *SQLiteRepository) GetSMTPSettings(userID int64) (*models.SMTPSettings, error) {
	query := `SELECT id, user_id, host, port, username, password, from_email, from_name, use_tls, updated_at
		FROM smtp_settings WHERE user_id = ?`

	settings := &models.SMTPSettings{}
	var fromName sql.NullString

	err := r.db.QueryRow(query, userID).Scan(&settings.ID, &settings.UserID, &settings.Host,
		&settings.Port, &settings.Username, &settings.Password, &settings.FromEmail,
		&fromName, &settings.UseTLS, &settings.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("smtp settings not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get smtp settings: %w", err)
	}

	if fromName.Valid {
		settings.FromName = fromName.String
	}

	return settings, nil
}

func (r *SQLiteRepository) UpsertSMTPSettings(settings *models.SMTPSettings) error {
	query := `INSERT INTO smtp_settings (user_id, host, port, username, password, from_email, from_name, use_tls)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
		host = excluded.host,
		port = excluded.port,
		username = excluded.username,
		password = excluded.password,
		from_email = excluded.from_email,
		from_name = excluded.from_name,
		use_tls = excluded.use_tls,
		updated_at = CURRENT_TIMESTAMP`

	result, err := r.db.Exec(query, settings.UserID, settings.Host, settings.Port,
		settings.Username, settings.Password, settings.FromEmail, settings.FromName, settings.UseTLS)
	if err != nil {
		return fmt.Errorf("failed to upsert smtp settings: %w", err)
	}

	if settings.ID == 0 {
		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get smtp settings ID: %w", err)
		}
		settings.ID = id
	}

	settings.UpdatedAt = time.Now()
	return nil
}
