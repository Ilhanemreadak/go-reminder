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

	// Email Log operations
	CreateEmailLog(log *models.EmailLog) (int64, error)
	UpdateEmailLogStatus(id int64, status string, errorMessage *string, sentAt time.Time) error
	GetEmailLog(id int64, userID int64) (*models.EmailLog, error)
	GetEmailLogs(filter *models.EmailHistoryFilter) (*models.EmailHistoryResult, error)

	// Recipient Group operations
	CreateRecipientGroup(group *models.RecipientGroup) error
	GetRecipientGroup(id int64) (*models.RecipientGroup, error)
	GetRecipientGroupsByUserID(userID int64) ([]*models.RecipientGroup, error)
	UpdateRecipientGroup(group *models.RecipientGroup) error
	DeleteRecipientGroup(id int64) error
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
		schedule_date, interval_days, day_of_week, day_of_month, time_of_day, last_sent_at, next_send_at, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query, reminder.UserID, reminder.Title, string(recipients),
		reminder.EmailContent, reminder.ScheduleType, reminder.ScheduleDate, reminder.IntervalDays, reminder.DayOfWeek,
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
	query := `SELECT id, user_id, title, recipients, email_content, schedule_type, schedule_date, interval_days,
		day_of_week, day_of_month, time_of_day, last_sent_at, next_send_at, is_active, created_at, updated_at
		FROM reminders WHERE id = ?`

	reminder := &models.Reminder{}
	var recipientsJSON string
	var lastSentAt sql.NullTime
	var scheduleDate sql.NullString

	err := r.db.QueryRow(query, id).Scan(&reminder.ID, &reminder.UserID, &reminder.Title,
		&recipientsJSON, &reminder.EmailContent, &reminder.ScheduleType, &scheduleDate, &reminder.IntervalDays,
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

	if scheduleDate.Valid {
		reminder.ScheduleDate = scheduleDate.String
	}

	return reminder, nil
}

func (r *SQLiteRepository) GetRemindersByUserID(userID int64) ([]*models.Reminder, error) {
	query := `SELECT id, user_id, title, recipients, email_content, schedule_type, schedule_date, interval_days,
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
		var scheduleDate sql.NullString

		err := rows.Scan(&reminder.ID, &reminder.UserID, &reminder.Title, &recipientsJSON,
			&reminder.EmailContent, &reminder.ScheduleType, &scheduleDate, &reminder.IntervalDays,
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

		if scheduleDate.Valid {
			reminder.ScheduleDate = scheduleDate.String
		}

		reminders = append(reminders, reminder)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reminders: %w", err)
	}

	return reminders, nil
}

func (r *SQLiteRepository) GetDueReminders(now time.Time) ([]*models.Reminder, error) {
	query := `SELECT id, user_id, title, recipients, email_content, schedule_type, schedule_date, interval_days,
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
		var scheduleDate sql.NullString

		err := rows.Scan(&reminder.ID, &reminder.UserID, &reminder.Title, &recipientsJSON,
			&reminder.EmailContent, &reminder.ScheduleType, &scheduleDate, &reminder.IntervalDays,
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

		if scheduleDate.Valid {
			reminder.ScheduleDate = scheduleDate.String
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
		schedule_date = ?, interval_days = ?, day_of_week = ?, day_of_month = ?, time_of_day = ?, last_sent_at = ?,
		next_send_at = ?, is_active = ?, updated_at = ? WHERE id = ?`

	_, err = r.db.Exec(query, reminder.Title, string(recipients), reminder.EmailContent,
		reminder.ScheduleType, reminder.ScheduleDate, reminder.IntervalDays, reminder.DayOfWeek, reminder.DayOfMonth,
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

// Email Log operations

func (r *SQLiteRepository) CreateEmailLog(log *models.EmailLog) (int64, error) {
	recipients, err := json.Marshal(log.Recipients)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal recipients: %w", err)
	}

	query := `INSERT INTO email_logs (reminder_id, user_id, recipients, subject, email_content, 
		status, error_message, sent_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query, log.ReminderID, log.UserID, string(recipients),
		log.Subject, log.EmailContent, log.Status, log.ErrorMessage, log.SentAt, time.Now())
	if err != nil {
		return 0, fmt.Errorf("failed to create email log: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get email log ID: %w", err)
	}

	return id, nil
}

func (r *SQLiteRepository) UpdateEmailLogStatus(id int64, status string, errorMessage *string, sentAt time.Time) error {
	query := `UPDATE email_logs SET status = ?, error_message = ?, sent_at = ? WHERE id = ?`

	_, err := r.db.Exec(query, status, errorMessage, sentAt, id)
	if err != nil {
		return fmt.Errorf("failed to update email log status: %w", err)
	}

	return nil
}

func (r *SQLiteRepository) GetEmailLog(id int64, userID int64) (*models.EmailLog, error) {
	query := `SELECT el.id, el.reminder_id, el.user_id, el.recipients, el.subject, el.email_content,
		el.status, el.error_message, el.sent_at, el.created_at, r.title
		FROM email_logs el
		LEFT JOIN reminders r ON el.reminder_id = r.id
		WHERE el.id = ? AND el.user_id = ?`

	log := &models.EmailLog{}
	var recipientsJSON string
	var reminderID sql.NullInt64
	var errorMessage sql.NullString
	var reminderTitle sql.NullString

	err := r.db.QueryRow(query, id, userID).Scan(
		&log.ID, &reminderID, &log.UserID, &recipientsJSON, &log.Subject,
		&log.EmailContent, &log.Status, &errorMessage, &log.SentAt,
		&log.CreatedAt, &reminderTitle)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("email log not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get email log: %w", err)
	}

	if err := json.Unmarshal([]byte(recipientsJSON), &log.Recipients); err != nil {
		return nil, fmt.Errorf("failed to unmarshal recipients: %w", err)
	}

	if reminderID.Valid {
		log.ReminderID = &reminderID.Int64
	}

	if errorMessage.Valid {
		log.ErrorMessage = &errorMessage.String
	}

	if reminderTitle.Valid {
		log.ReminderTitle = reminderTitle.String
	}

	return log, nil
}

func (r *SQLiteRepository) GetEmailLogs(filter *models.EmailHistoryFilter) (*models.EmailHistoryResult, error) {
	// Build WHERE clause dynamically based on filter parameters
	whereConditions := []string{"el.user_id = ?"}
	args := []interface{}{filter.UserID}

	if filter.ReminderID != nil {
		whereConditions = append(whereConditions, "el.reminder_id = ?")
		args = append(args, *filter.ReminderID)
	}

	if filter.Status != "" {
		whereConditions = append(whereConditions, "el.status = ?")
		args = append(args, filter.Status)
	}

	if filter.StartDate != nil {
		whereConditions = append(whereConditions, "el.sent_at >= ?")
		args = append(args, *filter.StartDate)
	}

	if filter.EndDate != nil {
		whereConditions = append(whereConditions, "el.sent_at <= ?")
		args = append(args, *filter.EndDate)
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + whereConditions[0]
		for i := 1; i < len(whereConditions); i++ {
			whereClause += " AND " + whereConditions[i]
		}
	}

	// Get total count
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM email_logs el %s`, whereClause)
	var totalCount int
	err := r.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count email logs: %w", err)
	}

	// Calculate pagination
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	// Get paginated logs with LEFT JOIN
	query := fmt.Sprintf(`SELECT el.id, el.reminder_id, el.user_id, el.recipients, el.subject, 
		el.email_content, el.status, el.error_message, el.sent_at, el.created_at, r.title
		FROM email_logs el
		LEFT JOIN reminders r ON el.reminder_id = r.id
		%s
		ORDER BY el.sent_at DESC
		LIMIT ? OFFSET ?`, whereClause)

	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get email logs: %w", err)
	}
	defer rows.Close()

	var logs []*models.EmailLog
	for rows.Next() {
		log := &models.EmailLog{}
		var recipientsJSON string
		var reminderID sql.NullInt64
		var errorMessage sql.NullString
		var reminderTitle sql.NullString

		err := rows.Scan(&log.ID, &reminderID, &log.UserID, &recipientsJSON, &log.Subject,
			&log.EmailContent, &log.Status, &errorMessage, &log.SentAt,
			&log.CreatedAt, &reminderTitle)
		if err != nil {
			return nil, fmt.Errorf("failed to scan email log: %w", err)
		}

		if err := json.Unmarshal([]byte(recipientsJSON), &log.Recipients); err != nil {
			return nil, fmt.Errorf("failed to unmarshal recipients: %w", err)
		}

		if reminderID.Valid {
			log.ReminderID = &reminderID.Int64
		}

		if errorMessage.Valid {
			log.ErrorMessage = &errorMessage.String
		}

		if reminderTitle.Valid {
			log.ReminderTitle = reminderTitle.String
		}

		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating email logs: %w", err)
	}

	result := &models.EmailHistoryResult{
		Logs:       logs,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	return result, nil
}

// Recipient Group operations

func (r *SQLiteRepository) CreateRecipientGroup(group *models.RecipientGroup) error {
	emails, err := json.Marshal(group.Emails)
	if err != nil {
		return fmt.Errorf("failed to marshal emails: %w", err)
	}

	query := `INSERT INTO recipient_groups (user_id, name, emails) VALUES (?, ?, ?)`
	result, err := r.db.Exec(query, group.UserID, group.Name, string(emails))
	if err != nil {
		return fmt.Errorf("failed to create recipient group: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get recipient group ID: %w", err)
	}
	group.ID = id
	group.CreatedAt = time.Now()
	group.UpdatedAt = time.Now()
	return nil
}

func (r *SQLiteRepository) GetRecipientGroup(id int64) (*models.RecipientGroup, error) {
	query := `SELECT id, user_id, name, emails, created_at, updated_at FROM recipient_groups WHERE id = ?`

	group := &models.RecipientGroup{}
	var emailsJSON string

	err := r.db.QueryRow(query, id).Scan(&group.ID, &group.UserID, &group.Name,
		&emailsJSON, &group.CreatedAt, &group.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("recipient group not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get recipient group: %w", err)
	}

	if err := json.Unmarshal([]byte(emailsJSON), &group.Emails); err != nil {
		return nil, fmt.Errorf("failed to unmarshal emails: %w", err)
	}

	return group, nil
}

func (r *SQLiteRepository) GetRecipientGroupsByUserID(userID int64) ([]*models.RecipientGroup, error) {
	query := `SELECT id, user_id, name, emails, created_at, updated_at 
		FROM recipient_groups WHERE user_id = ? ORDER BY name ASC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipient groups: %w", err)
	}
	defer rows.Close()

	var groups []*models.RecipientGroup
	for rows.Next() {
		group := &models.RecipientGroup{}
		var emailsJSON string

		err := rows.Scan(&group.ID, &group.UserID, &group.Name,
			&emailsJSON, &group.CreatedAt, &group.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recipient group: %w", err)
		}

		if err := json.Unmarshal([]byte(emailsJSON), &group.Emails); err != nil {
			return nil, fmt.Errorf("failed to unmarshal emails: %w", err)
		}

		groups = append(groups, group)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating recipient groups: %w", err)
	}

	return groups, nil
}

func (r *SQLiteRepository) UpdateRecipientGroup(group *models.RecipientGroup) error {
	emails, err := json.Marshal(group.Emails)
	if err != nil {
		return fmt.Errorf("failed to marshal emails: %w", err)
	}

	query := `UPDATE recipient_groups SET name = ?, emails = ?, updated_at = ? WHERE id = ?`
	_, err = r.db.Exec(query, group.Name, string(emails), time.Now(), group.ID)
	if err != nil {
		return fmt.Errorf("failed to update recipient group: %w", err)
	}

	group.UpdatedAt = time.Now()
	return nil
}

func (r *SQLiteRepository) DeleteRecipientGroup(id int64) error {
	query := `DELETE FROM recipient_groups WHERE id = ?`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete recipient group: %w", err)
	}
	return nil
}
