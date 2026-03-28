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

	// GenerateOpaqueToken creates a random opaque access token (raw, hash, prefix).
	GenerateOpaqueToken() (raw, hash, prefix string, err error)

	// HashToken creates a SHA-256 hash of a token for storage.
	HashToken(token string) string
}

// AccessToken represents a long-lived opaque access token for API access.
type AccessToken struct {
	ID        string
	UserID    string
	Name      string
	TokenHash string
	Prefix    string
	Status    string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// AccessTokenRepository defines the persistence operations for access tokens.
type AccessTokenRepository interface {
	// Create stores a new access token.
	Create(ctx context.Context, userID, name, tokenHash, prefix string, expiresAt time.Time) (*AccessToken, error)

	// GetByTokenHash retrieves an access token by its hash.
	GetByTokenHash(ctx context.Context, tokenHash string) (*AccessToken, error)

	// ListByUserID returns all access tokens for a user.
	ListByUserID(ctx context.Context, userID string) ([]AccessToken, error)

	// UpdateStatus changes the status of an access token.
	UpdateStatus(ctx context.Context, id, status string) error

	// MarkExpiredBatch sets status to 'expired' for the given token IDs.
	MarkExpiredBatch(ctx context.Context, ids []string) error

	// LoadAllActive returns all active access tokens.
	LoadAllActive(ctx context.Context) ([]AccessToken, error)
}

// AccessTokenStore defines the in-memory cache for fast token lookups.
type AccessTokenStore interface {
	// Lookup checks if a raw token is valid and returns the associated user ID.
	Lookup(rawToken string) (userID string, ok bool)

	// Add inserts a token entry into the in-memory store.
	Add(tokenHash, userID string, expiresAt time.Time)

	// Revoke marks a token as revoked in the in-memory store.
	Revoke(tokenHash string)

	// RemoveExpired purges expired entries from the in-memory store.
	RemoveExpired()
}

// PasswordHasher defines the interface for password hashing and verification.
type PasswordHasher interface {
	// Hash creates a bcrypt hash of the password.
	Hash(password string) (string, error)

	// Compare checks if a password matches a hash.
	Compare(hash, password string) error
}
