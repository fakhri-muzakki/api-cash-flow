package service

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"time"

	"cash-flow/internal/model"
	"cash-flow/internal/repository"
)

type TransactionService interface {
	Create(ctx context.Context, userID string, req *model.CreateTransactionRequest) (*model.Transaction, error)
	GetAll(ctx context.Context, userID string, filter repository.TransactionFilter) (*model.TransactionListResponse, error) // update ini
	GetByID(ctx context.Context, id, userID string) (*model.Transaction, error)
	Update(ctx context.Context, id, userID string, req *model.UpdateTransactionRequest) (*model.Transaction, error)
	Delete(ctx context.Context, id, userID string) error
}

type transactionService struct {
	txRepo   repository.TransactionRepository
	userRepo repository.UserRepository
	db       *sql.DB
}

func NewTransactionService(
	txRepo repository.TransactionRepository,
	userRepo repository.UserRepository,
	db *sql.DB,
) TransactionService {
	return &transactionService{
		txRepo:   txRepo,
		userRepo: userRepo,
		db:       db,
	}
}

func (s *transactionService) Create(ctx context.Context, userID string, req *model.CreateTransactionRequest) (*model.Transaction, error) {
	parsedDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, errors.New("invalid date format, use YYYY-MM-DD")
	}

	// Kalau expense, amount jadi negatif untuk update balance
	balanceDelta := req.Amount
	if req.Type == model.TransactionExpense {
		balanceDelta = -req.Amount
	}

	// Mulai DB transaction — create transaksi + update balance harus atomic
	tx, err := s.txRepo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	t := &model.Transaction{
		UserID: userID,
		Type:   req.Type,
		Amount: req.Amount,
		Note:   req.Note,
		Date:   parsedDate,
	}

	if err = s.txRepo.Create(ctx, tx, t); err != nil {
		return nil, err
	}

	if err = s.userRepo.UpdateBalance(ctx, tx, userID, balanceDelta); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return t, nil
}

func (s *transactionService) GetAll(ctx context.Context, userID string, filter repository.TransactionFilter) (*model.TransactionListResponse, error) {
	// Sanitasi nilai page dan limit
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 10 // default 10, maksimal 100
	}

	// Jalankan query data dan count secara bersamaan
	type countResult struct {
		total int64
		err   error
	}

	countCh := make(chan countResult, 1)
	go func() {
		total, err := s.txRepo.CountByUserID(ctx, userID, filter)
		countCh <- countResult{total, err}
	}()

	transactions, err := s.txRepo.FindByUserID(ctx, userID, filter)
	if err != nil {
		return nil, err
	}

	cr := <-countCh
	if cr.err != nil {
		return nil, cr.err
	}

	if transactions == nil {
		transactions = []*model.Transaction{}
	}

	totalPages := int(math.Ceil(float64(cr.total) / float64(filter.Limit)))
	if totalPages == 0 {
		totalPages = 1
	}

	return &model.TransactionListResponse{
		Data: func() []model.Transaction {
			result := make([]model.Transaction, len(transactions))
			for i, t := range transactions {
				result[i] = *t
			}
			return result
		}(),
		Meta: model.PaginationMeta{
			Page:       filter.Page,
			Limit:      filter.Limit,
			TotalItems: cr.total,
			TotalPages: totalPages,
		},
	}, nil
}

func (s *transactionService) GetByID(ctx context.Context, id, userID string) (*model.Transaction, error) {
	t, err := s.txRepo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, errors.New("transaction not found")
	}
	return t, nil
}

func (s *transactionService) Update(ctx context.Context, id, userID string, req *model.UpdateTransactionRequest) (*model.Transaction, error) {
	// Ambil data lama dulu
	existing, err := s.txRepo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("transaction not found")
	}

	// Hitung selisih balance:
	// Balikkan efek transaksi lama, lalu terapkan yang baru
	oldDelta := existing.Amount
	if existing.Type == model.TransactionExpense {
		oldDelta = -existing.Amount
	}

	// Apply perubahan dari request
	if req.Amount > 0 {
		existing.Amount = req.Amount
	}
	if req.Note != "" {
		existing.Note = req.Note
	}
	if req.Date != "" {
		parsedDate, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			return nil, errors.New("invalid date format, use YYYY-MM-DD")
		}
		existing.Date = parsedDate
	}

	newDelta := existing.Amount
	if existing.Type == model.TransactionExpense {
		newDelta = -existing.Amount
	}

	// Selisih yang perlu diaplikasikan ke balance
	balanceDiff := newDelta - oldDelta

	tx, err := s.txRepo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if err = s.txRepo.Update(ctx, tx, existing); err != nil {
		return nil, err
	}

	if balanceDiff != 0 {
		if err = s.userRepo.UpdateBalance(ctx, tx, userID, balanceDiff); err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return existing, nil
}

func (s *transactionService) Delete(ctx context.Context, id, userID string) error {
	// Ambil dulu untuk tahu amount-nya — perlu untuk reverse balance
	existing, err := s.txRepo.FindByID(ctx, id, userID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("transaction not found")
	}

	// Balikkan efek transaksi yang dihapus
	balanceDelta := -existing.Amount
	if existing.Type == model.TransactionExpense {
		balanceDelta = existing.Amount
	}

	tx, err := s.txRepo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	deleted, err := s.txRepo.Delete(ctx, tx, id, userID)
	if err != nil {
		return err
	}
	if deleted == nil {
		return errors.New("transaction not found")
	}

	if err = s.userRepo.UpdateBalance(ctx, tx, userID, balanceDelta); err != nil {
		return err
	}

	return tx.Commit()
}
