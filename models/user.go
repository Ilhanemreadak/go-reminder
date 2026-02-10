package models

import "time"

// User represents a system user
type User struct {
	ID           int64
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}
