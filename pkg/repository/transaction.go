package repository

import (
	"context"
	"database/sql"
	"fmt"

	"cash-flow/pkg/model"
)

type TransactionRepository interface {
	Create(ctx context.Context, tx *sql.Tx, t *model.Transaction) error
	FindByID(ctx context.Context, id, userID string) (*model.Transaction, error)
	FindByUserID(ctx context.Context, userID string, filter TransactionFilter) ([]*model.Transaction, error)
	CountByUserID(ctx context.Context, userID string, filter TransactionFilter) (int64, error) // tambah ini
	Update(ctx context.Context, tx *sql.Tx, t *model.Transaction) error
	Delete(ctx context.Context, tx *sql.Tx, id, userID string) (*model.Transaction, error)
	BeginTx(ctx context.Context) (*sql.Tx, error)
}

type TransactionFilter struct {
	Period    string // today, week, month, year, custom
	DateStart string
	DateEnd   string
	Page      int
	Limit     int
}

type transactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

func (r *transactionRepository) Create(ctx context.Context, tx *sql.Tx, t *model.Transaction) error {
	query := `
		INSERT INTO transactions (id, user_id, type, amount, note, date, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, now(), now())
		RETURNING id, created_at, updated_at
	`
	return tx.QueryRowContext(ctx, query, t.UserID, t.Type, t.Amount, t.Note, t.Date).
		Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}

func (r *transactionRepository) FindByID(ctx context.Context, id, userID string) (*model.Transaction, error) {
	t := &model.Transaction{}
	query := `SELECT id, user_id, type, amount, note, date, created_at, updated_at FROM transactions WHERE id = $1 AND user_id = $2`
	err := r.db.QueryRowContext(ctx, query, id, userID).
		Scan(&t.ID, &t.UserID, &t.Type, &t.Amount, &t.Note, &t.Date, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return t, err
}

func (r *transactionRepository) FindByUserID(ctx context.Context, userID string, filter TransactionFilter) ([]*model.Transaction, error) {
	query := `
		SELECT id, user_id, type, amount, note, date, created_at, updated_at
		FROM transactions
		WHERE user_id = $1
	`
	args := []any{userID}
	idx := 2

	query, args, idx = applyPeriodFilter(query, args, idx, filter)

	query += ` ORDER BY date DESC, created_at DESC`

	// Pagination
	query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, idx, idx+1)
	args = append(args, filter.Limit, (filter.Page-1)*filter.Limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*model.Transaction
	for rows.Next() {
		t := &model.Transaction{}
		if err := rows.Scan(&t.ID, &t.UserID, &t.Type, &t.Amount, &t.Note, &t.Date, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}
	return transactions, rows.Err()
}

func (r *transactionRepository) Update(ctx context.Context, tx *sql.Tx, t *model.Transaction) error {
	query := `
		UPDATE transactions SET amount = $1, note = $2, date = $3, updated_at = now()
		WHERE id = $4 AND user_id = $5
	`
	_, err := tx.ExecContext(ctx, query, t.Amount, t.Note, t.Date, t.ID, t.UserID)
	return err
}

func (r *transactionRepository) Delete(ctx context.Context, tx *sql.Tx, id, userID string) (*model.Transaction, error) {
	t := &model.Transaction{}
	query := `DELETE FROM transactions WHERE id = $1 AND user_id = $2 RETURNING id, type, amount`
	err := tx.QueryRowContext(ctx, query, id, userID).Scan(&t.ID, &t.Type, &t.Amount)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return t, err
}

func (r *transactionRepository) CountByUserID(ctx context.Context, userID string, filter TransactionFilter) (int64, error) {
	query := `SELECT COUNT(*) FROM transactions WHERE user_id = $1`
	args := []any{userID}
	idx := 2

	query, args, _ = applyPeriodFilter(query, args, idx, filter)

	var total int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&total)
	return total, err
}

// Helper — pisahkan filter period agar tidak duplikasi kode
func applyPeriodFilter(query string, args []any, idx int, filter TransactionFilter) (string, []any, int) {
	switch filter.Period {
	case "today":
		query += ` AND date = CURRENT_DATE`
	case "week":
		query += ` AND date >= date_trunc('week', CURRENT_DATE)`
	case "month":
		query += ` AND date >= date_trunc('month', CURRENT_DATE)`
	case "year":
		query += ` AND date >= date_trunc('year', CURRENT_DATE)`
	case "custom":
		if filter.DateStart != "" && filter.DateEnd != "" {
			query += fmt.Sprintf(` AND date BETWEEN $%d AND $%d`, idx, idx+1)
			args = append(args, filter.DateStart, filter.DateEnd)
			idx += 2
		}
	}
	return query, args, idx
}
