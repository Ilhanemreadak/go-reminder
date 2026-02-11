package models

import "time"

// RecipientGroup represents a named group of email recipients
type RecipientGroup struct {
	ID        int64
	UserID    int64
	Name      string
	Emails    []string // Email addresses in the group
	CreatedAt time.Time
	UpdatedAt time.Time
}
