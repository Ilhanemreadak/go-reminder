package models

import "time"

// SMTPSettings represents email server configuration
type SMTPSettings struct {
	ID        int64
	UserID    int64
	Host      string
	Port      int
	Username  string
	Password  string // Encrypted
	FromEmail string
	FromName  string
	UseTLS    bool
	UpdatedAt time.Time
}
