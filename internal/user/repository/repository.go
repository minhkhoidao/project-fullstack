package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kyle/product/internal/user/model"
)

var (
	ErrNotFound       = errors.New("user: not found")
	ErrDuplicateEmail = errors.New("user: duplicate email")
)

// UserRepository defines persistence operations for users and addresses.
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	CreateAddress(ctx context.Context, addr *model.Address) error
	ListAddresses(ctx context.Context, userID string) ([]model.Address, error)
	DeleteAddress(ctx context.Context, id, userID string) error
}

// Compile-time interface compliance check.
var _ UserRepository = (*pgRepo)(nil)

type pgRepo struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository returns a UserRepository backed by PostgreSQL.
func NewPostgresRepository(pool *pgxpool.Pool) UserRepository {
	return &pgRepo{pool: pool}
}

func (r *pgRepo) Create(ctx context.Context, user *model.User) error {
	const q = `
		INSERT INTO users.users (id, email, password_hash, first_name, last_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.pool.Exec(ctx, q,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrDuplicateEmail
		}
		return fmt.Errorf("repo create user: %w", err)
	}
	return nil
}

func (r *pgRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
	const q = `
		SELECT id, email, password_hash, first_name, last_name, role, created_at, updated_at
		FROM users.users
		WHERE id = $1`

	var u model.User
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash,
		&u.FirstName, &u.LastName, &u.Role,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repo get user by id: %w", err)
	}
	return &u, nil
}

func (r *pgRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	const q = `
		SELECT id, email, password_hash, first_name, last_name, role, created_at, updated_at
		FROM users.users
		WHERE email = $1`

	var u model.User
	err := r.pool.QueryRow(ctx, q, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash,
		&u.FirstName, &u.LastName, &u.Role,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repo get user by email: %w", err)
	}
	return &u, nil
}

func (r *pgRepo) Update(ctx context.Context, user *model.User) error {
	const q = `
		UPDATE users.users
		SET first_name = $1, last_name = $2, updated_at = $3
		WHERE id = $4`

	tag, err := r.pool.Exec(ctx, q, user.FirstName, user.LastName, user.UpdatedAt, user.ID)
	if err != nil {
		return fmt.Errorf("repo update user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *pgRepo) CreateAddress(ctx context.Context, addr *model.Address) error {
	const q = `
		INSERT INTO users.addresses (id, user_id, label, street, city, state, zip, country, is_default, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.pool.Exec(ctx, q,
		addr.ID,
		addr.UserID,
		addr.Label,
		addr.Street,
		addr.City,
		addr.State,
		addr.Zip,
		addr.Country,
		addr.IsDefault,
		addr.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("repo create address: %w", err)
	}
	return nil
}

func (r *pgRepo) ListAddresses(ctx context.Context, userID string) ([]model.Address, error) {
	const q = `
		SELECT id, user_id, label, street, city, state, zip, country, is_default, created_at
		FROM users.addresses
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("repo list addresses: %w", err)
	}
	defer rows.Close()

	var addrs []model.Address
	for rows.Next() {
		var a model.Address
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.Label, &a.Street,
			&a.City, &a.State, &a.Zip, &a.Country,
			&a.IsDefault, &a.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("repo scan address: %w", err)
		}
		addrs = append(addrs, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo iterate addresses: %w", err)
	}
	return addrs, nil
}

func (r *pgRepo) DeleteAddress(ctx context.Context, id, userID string) error {
	const q = `DELETE FROM users.addresses WHERE id = $1 AND user_id = $2`

	tag, err := r.pool.Exec(ctx, q, id, userID)
	if err != nil {
		return fmt.Errorf("repo delete address: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
