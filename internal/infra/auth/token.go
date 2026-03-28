package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/emirakts0/mahzen/internal/config"
)

// TokenProvider implements domain.TokenGenerator using JWT for access tokens
// and random bytes for refresh tokens.
type TokenProvider struct {
	secret             []byte
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
}

// NewTokenProvider creates a new TokenProvider from auth config.
func NewTokenProvider(cfg config.AuthConfig) *TokenProvider {
	expiry := cfg.AccessTokenExpiry
	if expiry == 0 {
		expiry = 15 * time.Minute
	}
	refreshExpiry := cfg.RefreshTokenExpiry
	if refreshExpiry == 0 {
		refreshExpiry = 7 * 24 * time.Hour
	}
	return &TokenProvider{
		secret:             []byte(cfg.JWTSecret),
		accessTokenExpiry:  expiry,
		refreshTokenExpiry: refreshExpiry,
	}
}

// RefreshTokenExpiry returns the configured refresh token expiry duration.
func (p *TokenProvider) RefreshTokenExpiry() time.Duration {
	return p.refreshTokenExpiry
}

// GenerateAccessToken creates a short-lived JWT access token for a user.
func (p *TokenProvider) GenerateAccessToken(userID string) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(p.accessTokenExpiry)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(p.secret)
	if err != nil {
		return "", fmt.Errorf("signing access token: %w", err)
	}
	return signed, nil
}

// ValidateAccessToken validates a JWT access token and returns the user ID.
func (p *TokenProvider) ValidateAccessToken(tokenStr string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return p.secret, nil
	})
	if err != nil {
		return "", fmt.Errorf("parsing access token: %w", err)
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("invalid token claims")
	}

	return claims.Subject, nil
}

// GenerateRefreshToken creates a random refresh token string (32 bytes, hex-encoded).
func (p *TokenProvider) GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating refresh token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// HashToken creates a SHA-256 hash of a token for storage.
func (p *TokenProvider) HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// GenerateOpaqueToken creates a random opaque access token.
// Returns the raw token (prefixed with "mah_"), its SHA-256 hash, and a display prefix.
func (p *TokenProvider) GenerateOpaqueToken() (raw, hash, prefix string, err error) {
	b := make([]byte, 48)
	if _, err = rand.Read(b); err != nil {
		return "", "", "", fmt.Errorf("generating opaque token: %w", err)
	}
	raw = "mah_" + hex.EncodeToString(b)
	h := sha256.Sum256([]byte(raw))
	hash = hex.EncodeToString(h[:])
	prefix = raw[:12] + "..."
	return raw, hash, prefix, nil
}
