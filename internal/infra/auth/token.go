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

// TokenProvider implements domain.TokenGenerator with JWT access tokens and
// random-bytes refresh/opaque tokens.
type TokenProvider struct {
	secret             []byte
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
}

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
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
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
	return randomHex(32)
}

// HashToken creates a SHA-256 hash of a token for storage.
func (p *TokenProvider) HashToken(token string) string {
	return hex.EncodeToString(p.sha256(token))
}

// GenerateOpaqueToken creates a random opaque access token.
// Returns the raw token (prefixed with "mah_"), its SHA-256 hash, and a display prefix.
func (p *TokenProvider) GenerateOpaqueToken() (raw, hash, prefix string, err error) {
	b, err := randomHex(48)
	if err != nil {
		return "", "", "", err
	}
	raw = "mah_" + b
	hash = p.HashToken(raw)
	prefix = raw[:12] + "..."
	return raw, hash, prefix, nil
}

// randomHex returns n cryptographically random bytes, hex-encoded.
func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating random bytes: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func (p *TokenProvider) sha256(s string) []byte {
	h := sha256.Sum256([]byte(s))
	return h[:]
}
