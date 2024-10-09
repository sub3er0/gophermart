package interfaces

import (
	"gophermart/internal/models"
)

type UserBalanceRepositoryInterface interface {
	UpdateUserBalance(accrual float32, userID int) error
	GetUserBalance(userID int) (UserBalance, error)
	CreateUserBalance(user models.User) error
}

type UserBalance struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}
