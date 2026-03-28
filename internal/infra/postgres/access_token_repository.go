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

// AccessTokenRepository implements domain.AccessTokenRepository using PostgreSQL.
type AccessTokenRepository struct {
	q *query.Queries
}

// NewAccessTokenRepository creates a new AccessTokenRepository backed by the given connection pool.
func NewAccessTokenRepository(pool *pgxpool.Pool) *AccessTokenRepository {
	return &AccessTokenRepository{q: query.New(pool)}
}

func (r *AccessTokenRepository) Create(ctx context.Context, userID, name, tokenHash, prefix string, expiresAt time.Time) (*domain.AccessToken, error) {
	uid, err := parseUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("parsing user id: %w", err)
	}

	row, err := r.q.CreateAccessToken(ctx, query.CreateAccessTokenParams{
		UserID:    uid,
		Name:      name,
		TokenHash: tokenHash,
		Prefix:    prefix,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("creating access token: %w", err)
	}

	return mapAccessTokenRow(row), nil
}

func (r *AccessTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.AccessToken, error) {
	row, err := r.q.GetAccessTokenByHash(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("getting access token: %w", err)
	}

	return mapAccessTokenRow(row), nil
}

func (r *AccessTokenRepository) ListByUserID(ctx context.Context, userID string) ([]domain.AccessToken, error) {
	uid, err := parseUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("parsing user id: %w", err)
	}

	rows, err := r.q.ListAccessTokensByUserID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("listing access tokens: %w", err)
	}

	tokens := make([]domain.AccessToken, len(rows))
	for i, row := range rows {
		tokens[i] = *mapAccessTokenRow(row)
	}
	return tokens, nil
}

func (r *AccessTokenRepository) UpdateStatus(ctx context.Context, id, status string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return fmt.Errorf("parsing token id: %w", err)
	}

	if err := r.q.UpdateAccessTokenStatus(ctx, query.UpdateAccessTokenStatusParams{
		ID:     uid,
		Status: status,
	}); err != nil {
		return fmt.Errorf("updating access token status: %w", err)
	}
	return nil
}

func (r *AccessTokenRepository) MarkExpiredBatch(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	uuids := make([]pgtype.UUID, len(ids))
	for i, id := range ids {
		uid, err := parseUUID(id)
		if err != nil {
			return fmt.Errorf("parsing token id %q: %w", id, err)
		}
		uuids[i] = uid
	}

	if err := r.q.MarkAccessTokensExpiredBatch(ctx, uuids); err != nil {
		return fmt.Errorf("marking tokens expired: %w", err)
	}
	return nil
}

func (r *AccessTokenRepository) LoadAllActive(ctx context.Context) ([]domain.AccessToken, error) {
	rows, err := r.q.LoadAllActiveAccessTokens(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading active access tokens: %w", err)
	}

	tokens := make([]domain.AccessToken, len(rows))
	for i, row := range rows {
		tokens[i] = *mapAccessTokenRow(row)
	}
	return tokens, nil
}

func mapAccessTokenRow(row query.AccessToken) *domain.AccessToken {
	return &domain.AccessToken{
		ID:        uuidToString(row.ID),
		UserID:    uuidToString(row.UserID),
		Name:      row.Name,
		TokenHash: row.TokenHash,
		Prefix:    row.Prefix,
		Status:    row.Status,
		ExpiresAt: row.ExpiresAt.Time,
		CreatedAt: row.CreatedAt.Time,
	}
}
