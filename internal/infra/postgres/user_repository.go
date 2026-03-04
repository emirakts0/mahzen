package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/emirakts0/mahzen/internal/domain"
	"github.com/emirakts0/mahzen/internal/infra/postgres/query"
)

// UserRepository implements domain.UserRepository using PostgreSQL.
type UserRepository struct {
	q *query.Queries
}

// NewUserRepository creates a new UserRepository backed by the given connection pool.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{q: query.New(pool)}
}

func (r *UserRepository) Create(ctx context.Context, email, displayName, passwordHash string) (*domain.User, error) {
	row, err := r.q.CreateUser(ctx, query.CreateUserParams{
		Email:        email,
		DisplayName:  displayName,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	return mapUserRow(row), nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, fmt.Errorf("parsing user id: %w", err)
	}

	row, err := r.q.GetUserByID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("getting user by id: %w", err)
	}

	return mapUserRow(row), nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("getting user by email: %w", err)
	}

	return mapUserRow(row), nil
}

func mapUserRow(row query.User) *domain.User {
	return &domain.User{
		ID:           uuidToString(row.ID),
		Email:        row.Email,
		DisplayName:  row.DisplayName,
		PasswordHash: row.PasswordHash,
		CreatedAt:    row.CreatedAt.Time,
	}
}
