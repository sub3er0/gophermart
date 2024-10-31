package service

import (
	"gophermart/internal/interfaces"
	"gophermart/internal/models"
	"gophermart/internal/repository"

	"github.com/shopspring/decimal"
)

type UserBalanceService struct {
	UserBalanceRepository *repository.UserBalanceRepository
}

func (ubs *UserBalanceService) UpdateUserBalance(accrual decimal.Decimal, userID int) error {
	return ubs.UserBalanceRepository.UpdateUserBalance(accrual, userID)
}

func (ubs *UserBalanceService) GetUserBalance(userID int) (interfaces.UserBalance, error) {
	return ubs.UserBalanceRepository.GetUserBalance(userID)
}

func (ubs *UserBalanceService) CreateUserBalance(user models.User) error {
	return ubs.UserBalanceRepository.CreateUserBalance(user)
}
