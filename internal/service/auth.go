package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/emirakts0/mahzen/internal/domain"
)

// AuthTokens holds the access and refresh tokens returned after authentication.
type AuthTokens struct {
	AccessToken  string
	RefreshToken string
}

// AuthService handles registration, login, token refresh and logout.
type AuthService struct {
	userRepo   domain.UserRepository
	tokenRepo  domain.RefreshTokenRepository
	tokenGen   domain.TokenGenerator
	hasher     domain.PasswordHasher
	refreshTTL time.Duration
}

// NewAuthService creates a new AuthService.
func NewAuthService(
	userRepo domain.UserRepository,
	tokenRepo domain.RefreshTokenRepository,
	tokenGen domain.TokenGenerator,
	hasher domain.PasswordHasher,
	refreshTTL time.Duration,
) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		tokenGen:   tokenGen,
		hasher:     hasher,
		refreshTTL: refreshTTL,
	}
}

// Register creates a new user and returns auth tokens.
func (s *AuthService) Register(ctx context.Context, email, displayName, password string) (*AuthTokens, error) {
	slog.Info("registering user", "email", email, "display_name", displayName)

	hash, err := s.hasher.Hash(password)
	if err != nil {
		slog.Error("password hashing failed", "email", email, "error", err)
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	user, err := s.userRepo.Create(ctx, email, displayName, hash)
	if err != nil {
		slog.Error("user creation failed", "email", email, "error", err)
		return nil, fmt.Errorf("creating user: %w", err)
	}

	slog.Info("user registered", "user_id", user.ID, "email", email)
	return s.issueTokens(ctx, user.ID)
}

// Login authenticates a user with email and password, returns auth tokens.
func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthTokens, error) {
	slog.Info("login attempt", "email", email)

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		slog.Warn("login failed: user not found", "email", email)
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := s.hasher.Compare(user.PasswordHash, password); err != nil {
		slog.Warn("login failed: invalid password", "email", email, "user_id", user.ID)
		return nil, fmt.Errorf("invalid credentials")
	}

	slog.Info("login successful", "user_id", user.ID, "email", email)
	return s.issueTokens(ctx, user.ID)
}

// RefreshToken validates a refresh token and issues a new token pair.
// The old refresh token is consumed (deleted) and a new one is issued.
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*AuthTokens, error) {
	slog.Info("refreshing token")

	tokenHash := s.tokenGen.HashToken(refreshToken)

	stored, err := s.tokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		slog.Warn("token refresh failed: token not found")
		return nil, fmt.Errorf("invalid refresh token")
	}

	if time.Now().After(stored.ExpiresAt) {
		slog.Warn("token refresh failed: token expired", "user_id", stored.UserID, "expired_at", stored.ExpiresAt)
		_ = s.tokenRepo.DeleteByTokenHash(ctx, tokenHash)
		return nil, fmt.Errorf("refresh token expired")
	}

	// Consume the old refresh token (rotation).
	if err := s.tokenRepo.DeleteByTokenHash(ctx, tokenHash); err != nil {
		slog.Error("failed to consume refresh token", "user_id", stored.UserID, "error", err)
		return nil, fmt.Errorf("consuming refresh token: %w", err)
	}

	slog.Info("token refreshed", "user_id", stored.UserID)
	return s.issueTokens(ctx, stored.UserID)
}

// Logout invalidates the given refresh token.
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	slog.Info("logout request")

	tokenHash := s.tokenGen.HashToken(refreshToken)
	if err := s.tokenRepo.DeleteByTokenHash(ctx, tokenHash); err != nil {
		slog.Error("logout failed", "error", err)
		return fmt.Errorf("revoking refresh token: %w", err)
	}

	slog.Info("logout successful")
	return nil
}

// issueTokens generates an access token and a refresh token, stores the
// refresh token hash in the database, and returns the pair.
func (s *AuthService) issueTokens(ctx context.Context, userID string) (*AuthTokens, error) {
	accessToken, err := s.tokenGen.GenerateAccessToken(userID)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}

	refreshToken, err := s.tokenGen.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}

	tokenHash := s.tokenGen.HashToken(refreshToken)
	expiresAt := time.Now().Add(s.refreshTTL)

	if _, err := s.tokenRepo.Create(ctx, userID, tokenHash, expiresAt); err != nil {
		return nil, fmt.Errorf("storing refresh token: %w", err)
	}

	return &AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
