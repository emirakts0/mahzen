package domain

import (
	"context"
	"time"
)

// RefreshToken represents a stored refresh token for a user session.
type RefreshToken struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// RefreshTokenRepository defines the persistence operations for refresh tokens.
type RefreshTokenRepository interface {
	// Create stores a new refresh token.
	Create(ctx context.Context, userID, tokenHash string, expiresAt time.Time) (*RefreshToken, error)

	// GetByTokenHash retrieves a refresh token by its hash.
	GetByTokenHash(ctx context.Context, tokenHash string) (*RefreshToken, error)

	// DeleteByTokenHash removes a refresh token by its hash.
	DeleteByTokenHash(ctx context.Context, tokenHash string) error

	// DeleteAllForUser removes all refresh tokens for a given user.
	DeleteAllForUser(ctx context.Context, userID string) error
}

// TokenGenerator defines the interface for creating and validating JWT tokens.
type TokenGenerator interface {
	// GenerateAccessToken creates a short-lived JWT access token for a user.
	GenerateAccessToken(userID string) (string, error)

	// ValidateAccessToken validates a JWT access token and returns the user ID.
	ValidateAccessToken(token string) (string, error)

	// GenerateRefreshToken creates a random refresh token string.
	GenerateRefreshToken() (string, error)

	// HashToken creates a SHA-256 hash of a token for storage.
	HashToken(token string) string
}

// PasswordHasher defines the interface for password hashing and verification.
type PasswordHasher interface {
	// Hash creates a bcrypt hash of the password.
	Hash(password string) (string, error)

	// Compare checks if a password matches a hash.
	Compare(hash, password string) error
}
