package model

import "time"

type TransactionType string

const (
	TransactionIncome  TransactionType = "income"
	TransactionExpense TransactionType = "expense"
)

type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

type TransactionListResponse struct {
	Data []Transaction  `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

type Transaction struct {
	ID        string          `json:"id"`
	UserID    string          `json:"user_id"`
	Type      TransactionType `json:"type"`
	Amount    float64         `json:"amount"`
	Note      string          `json:"note"`
	Date      time.Time       `json:"date"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type CreateTransactionRequest struct {
	Type   TransactionType `json:"type" binding:"required,oneof=income expense"`
	Amount float64         `json:"amount" binding:"required,gt=0"`
	Note   string          `json:"note"`
	Date   string          `json:"date" binding:"required"` // format: 2006-01-02
}

type UpdateTransactionRequest struct {
	Amount float64 `json:"amount" binding:"omitempty,gt=0"`
	Note   string  `json:"note"`
	Date   string  `json:"date"`
}
