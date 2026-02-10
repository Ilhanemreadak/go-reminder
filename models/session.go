package models

import "time"

// Session represents an authenticated user session
type Session struct {
	ID        string
	UserID    int64
	ExpiresAt time.Time
	CreatedAt time.Time
}
