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

func mapUserRow[T query.CreateUserRow | query.GetUserByIDRow | query.GetUserByEmailRow](row T) *domain.User {
	// All three row types share identical fields; cast through any to access them.
	switch v := any(row).(type) {
	case query.CreateUserRow:
		return &domain.User{
			ID:           uuidToString(v.ID),
			Email:        v.Email,
			DisplayName:  v.DisplayName,
			PasswordHash: v.PasswordHash,
			CreatedAt:    v.CreatedAt.Time,
		}
	case query.GetUserByIDRow:
		return &domain.User{
			ID:           uuidToString(v.ID),
			Email:        v.Email,
			DisplayName:  v.DisplayName,
			PasswordHash: v.PasswordHash,
			CreatedAt:    v.CreatedAt.Time,
		}
	case query.GetUserByEmailRow:
		return &domain.User{
			ID:           uuidToString(v.ID),
			Email:        v.Email,
			DisplayName:  v.DisplayName,
			PasswordHash: v.PasswordHash,
			CreatedAt:    v.CreatedAt.Time,
		}
	default:
		return nil
	}
}
