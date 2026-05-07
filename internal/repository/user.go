package repository

import (
	"context"
	"database/sql"

	"cash-flow/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id string) (*model.User, error)
	UpdateBalance(ctx context.Context, tx *sql.Tx, userID string, amount float64) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (id, email, name, password, balance, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, 0, now(), now())
		RETURNING id, balance, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query, user.Email, user.Name, user.Password).
		Scan(&user.ID, &user.Balance, &user.CreatedAt, &user.UpdatedAt)
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, email, name, password, balance, created_at, updated_at FROM users WHERE email = $1`
	err := r.db.QueryRowContext(ctx, query, email).
		Scan(&user.ID, &user.Email, &user.Name, &user.Password, &user.Balance, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, email, name, password, balance, created_at, updated_at FROM users WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&user.ID, &user.Email, &user.Name, &user.Password, &user.Balance, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

// UpdateBalance dipanggil di dalam DB transaction — terima *sql.Tx bukan *sql.DB
func (r *userRepository) UpdateBalance(ctx context.Context, tx *sql.Tx, userID string, amount float64) error {
	query := `UPDATE users SET balance = balance + $1, updated_at = now() WHERE id = $2`
	_, err := tx.ExecContext(ctx, query, amount, userID)
	return err
}
