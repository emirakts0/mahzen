package domain

import "time"

// User represents an authenticated user of the platform.
type User struct {
	ID          string
	Email       string
	DisplayName string
	CreatedAt   time.Time
}
