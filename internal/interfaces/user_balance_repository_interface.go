package interfaces

import (
	"gophermart/internal/models"
	"gophermart/internal/repository"
)

type UserBalanceRepositoryInterface interface {
	UpdateUserBalance(accrual float32, userID int) error
	GetUserBalance(userID int) (repository.UserBalance, error)
	CreateUserBalance(user models.User) error
}
