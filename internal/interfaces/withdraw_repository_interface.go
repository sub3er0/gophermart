package interfaces

import (
	"time"
)

type WithdrawRepositoryInterface interface {
	Withdraw(userID int, orderNumber string, sum float32) (int, error)
	Withdrawals(userID int) ([]WithdrawInfo, error)
}

type WithdrawInfo struct {
	OrderNumber string    `json:"order"`
	Sum         float32   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}
