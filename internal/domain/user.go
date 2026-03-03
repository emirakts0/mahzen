package domain

import (
	"context"
	"time"
)

// User represents an authenticated user of the platform.
type User struct {
	ID           string
	Email        string
	DisplayName  string
	PasswordHash string
	CreatedAt    time.Time
}

// UserRepository defines the persistence operations for users.
type UserRepository interface {
	// Create inserts a new user and returns it.
	Create(ctx context.Context, email, displayName, passwordHash string) (*User, error)

	// GetByID retrieves a user by their internal ID.
	GetByID(ctx context.Context, id string) (*User, error)

	// GetByEmail retrieves a user by their email address.
	GetByEmail(ctx context.Context, email string) (*User, error)
}
