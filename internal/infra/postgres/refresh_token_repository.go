package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/emirakts0/mahzen/internal/domain"
	"github.com/emirakts0/mahzen/internal/infra/postgres/query"
)

// RefreshTokenRepository implements domain.RefreshTokenRepository using PostgreSQL.
type RefreshTokenRepository struct {
	q *query.Queries
}

// NewRefreshTokenRepository creates a new RefreshTokenRepository backed by the given connection pool.
func NewRefreshTokenRepository(pool *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{q: query.New(pool)}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, userID, tokenHash string, expiresAt time.Time) (*domain.RefreshToken, error) {
	uid, err := parseUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("parsing user id: %w", err)
	}

	row, err := r.q.CreateRefreshToken(ctx, query.CreateRefreshTokenParams{
		UserID:    uid,
		TokenHash: tokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("creating refresh token: %w", err)
	}

	return mapRefreshTokenRow(row), nil
}

func (r *RefreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	row, err := r.q.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("getting refresh token: %w", err)
	}

	return mapRefreshTokenRow(row), nil
}

func (r *RefreshTokenRepository) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	if err := r.q.DeleteRefreshTokenByHash(ctx, tokenHash); err != nil {
		return fmt.Errorf("deleting refresh token: %w", err)
	}
	return nil
}

func (r *RefreshTokenRepository) DeleteAllForUser(ctx context.Context, userID string) error {
	uid, err := parseUUID(userID)
	if err != nil {
		return fmt.Errorf("parsing user id: %w", err)
	}

	if err := r.q.DeleteRefreshTokensByUserID(ctx, uid); err != nil {
		return fmt.Errorf("deleting refresh tokens for user: %w", err)
	}
	return nil
}

func mapRefreshTokenRow(row query.RefreshToken) *domain.RefreshToken {
	return &domain.RefreshToken{
		ID:        uuidToString(row.ID),
		UserID:    uuidToString(row.UserID),
		TokenHash: row.TokenHash,
		ExpiresAt: row.ExpiresAt.Time,
		CreatedAt: row.CreatedAt.Time,
	}
}
