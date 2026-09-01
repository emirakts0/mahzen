package domain

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// User represents an authenticated user of the platform.
type User struct {
	ID           string
	Username     string
	Email        string
	DisplayName  string
	PasswordHash string
	CreatedAt    time.Time
}

// usernamePattern is the allowed username format: 3–32 chars, lowercase
// letters, digits and underscores, starting with a letter.
var usernamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{2,31}$`)

// NormalizeUsername trims and lowercases a username candidate.
func NormalizeUsername(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

// ValidateUsername reports whether the username matches the allowed format.
func ValidateUsername(username string) error {
	if !usernamePattern.MatchString(username) {
		return fmt.Errorf("username must be 3-32 characters: lowercase letters, digits and underscores, starting with a letter")
	}
	return nil
}

// UserRepository defines the persistence operations for users.
type UserRepository interface {
	// Create inserts a new user and returns it.
	Create(ctx context.Context, username, email, displayName, passwordHash string) (*User, error)

	// GetByID retrieves a user by their internal ID.
	GetByID(ctx context.Context, id string) (*User, error)

	// GetByEmail retrieves a user by their email address.
	GetByEmail(ctx context.Context, email string) (*User, error)

	// GetByEmailOrUsername retrieves a user by email address or username.
	GetByEmailOrUsername(ctx context.Context, identifier string) (*User, error)
}
