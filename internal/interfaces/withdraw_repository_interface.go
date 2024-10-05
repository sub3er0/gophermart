package interfaces

import (
	"gophermart/internal/repository"
)

type WithdrawRepositoryInterface interface {
	Withdraw(userID int, orderNumber string, sum float32) (int, error)
	Withdrawals(userID int) ([]repository.WithdrawInfo, error)
}
