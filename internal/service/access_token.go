package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/emirakts0/mahzen/internal/domain"
)

// AccessTokenService handles access token lifecycle.
type AccessTokenService struct {
	tokenRepo  domain.AccessTokenRepository
	tokenGen   domain.TokenGenerator
	tokenStore domain.AccessTokenStore
	defaultTTL time.Duration
}

// NewAccessTokenService creates a new AccessTokenService.
func NewAccessTokenService(
	tokenRepo domain.AccessTokenRepository,
	tokenGen domain.TokenGenerator,
	tokenStore domain.AccessTokenStore,
	defaultTTL time.Duration,
) *AccessTokenService {
	return &AccessTokenService{
		tokenRepo:  tokenRepo,
		tokenGen:   tokenGen,
		tokenStore: tokenStore,
		defaultTTL: defaultTTL,
	}
}

// Store returns the underlying in-memory token store for middleware use.
func (s *AccessTokenService) Store() domain.AccessTokenStore {
	return s.tokenStore
}

// CreateToken generates, persists, and caches a new access token.
func (s *AccessTokenService) CreateToken(ctx context.Context, userID, name string, expiresIn *time.Duration) (*domain.AccessToken, string, error) {
	ttl := s.defaultTTL
	if expiresIn != nil && *expiresIn > 0 {
		ttl = *expiresIn
	}

	raw, hash, prefix, err := s.tokenGen.GenerateOpaqueToken()
	if err != nil {
		return nil, "", fmt.Errorf("generating opaque token: %w", err)
	}

	expiresAt := time.Now().Add(ttl)

	token, err := s.tokenRepo.Create(ctx, userID, name, hash, prefix, expiresAt)
	if err != nil {
		return nil, "", fmt.Errorf("persisting access token: %w", err)
	}

	s.tokenStore.Add(hash, userID, expiresAt)

	slog.Info("access token created", "user_id", userID, "token_id", token.ID, "name", name)
	return token, raw, nil
}

// ListTokens returns all access tokens for a user.
func (s *AccessTokenService) ListTokens(ctx context.Context, userID string) ([]domain.AccessToken, error) {
	return s.tokenRepo.ListByUserID(ctx, userID)
}

// RevokeToken revokes an access token after verifying ownership.
func (s *AccessTokenService) RevokeToken(ctx context.Context, userID, tokenID string) error {
	tokens, err := s.tokenRepo.ListByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("listing tokens: %w", err)
	}

	var tokenHash string
	for _, t := range tokens {
		if t.ID == tokenID {
			tokenHash = t.TokenHash
			break
		}
	}
	if tokenHash == "" {
		return fmt.Errorf("token not found")
	}

	if err := s.tokenRepo.UpdateStatus(ctx, tokenID, "revoked"); err != nil {
		return fmt.Errorf("revoking token in db: %w", err)
	}

	s.tokenStore.Revoke(tokenHash)

	slog.Info("access token revoked", "user_id", userID, "token_id", tokenID)
	return nil
}

// LoadCacheFromDB loads all active tokens from the database into the in-memory store.
func (s *AccessTokenService) LoadCacheFromDB(ctx context.Context) error {
	tokens, err := s.tokenRepo.LoadAllActive(ctx)
	if err != nil {
		return fmt.Errorf("loading active tokens: %w", err)
	}

	for _, t := range tokens {
		s.tokenStore.Add(t.TokenHash, t.UserID, t.ExpiresAt)
	}

	slog.Info("access token cache loaded", "count", len(tokens))
	return nil
}

// StartExpiryWorker starts a background goroutine that periodically cleans up
// expired tokens from both the in-memory store and the database.
func (s *AccessTokenService) StartExpiryWorker(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				s.removeExpired(ctx)
			}
		}
	}()
}

func (s *AccessTokenService) removeExpired(ctx context.Context) {
	s.tokenStore.RemoveExpired()

	// Find active tokens in DB that are past their expiry.
	tokens, err := s.tokenRepo.LoadAllActive(ctx)
	if err != nil {
		slog.Error("failed to load active tokens for expiry check", "error", err)
		return
	}

	now := time.Now()
	var expiredIDs []string
	for _, t := range tokens {
		if now.After(t.ExpiresAt) {
			expiredIDs = append(expiredIDs, t.ID)
		}
	}

	if len(expiredIDs) > 0 {
		if err := s.tokenRepo.MarkExpiredBatch(ctx, expiredIDs); err != nil {
			slog.Error("failed to mark tokens expired", "count", len(expiredIDs), "error", err)
		} else {
			slog.Info("marked access tokens as expired", "count", len(expiredIDs))
		}
	}
}
