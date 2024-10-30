package service

import (
	"fmt"
	"gophermart/internal/interfaces"
	"gophermart/internal/repository"

	"github.com/shopspring/decimal"
)

type WithdrawService struct {
	WithdrawRepository repository.WithdrawRepositoryInterface
}

func (ws *WithdrawService) Withdraw(userID int, orderNumber string, sum decimal.Decimal) (int, error) {
	wr := ws.WithdrawRepository
	err := wr.GetDBStorage().BeginTransaction()
	if err != nil {
		return repository.NotEnoughFound, fmt.Errorf("ошибка при начале транзакции: %v", err)
	}
	defer func() {
		if r := recover(); r != nil {
			_ = wr.GetDBStorage().Rollback()
		}
	}()

	userBalance, err := wr.GetCurrentBalance(userID)
	if err != nil {
		_ = wr.GetDBStorage().Rollback()
		return repository.WithdrawTransactionError, err
	}

	if userBalance.LessThan(sum) {
		_ = wr.GetDBStorage().Rollback()
		return repository.NotEnoughFound, nil
	}

	if err = wr.UpdateUserBalance(userID, sum); err != nil {
		_ = wr.GetDBStorage().Rollback()
		return repository.WithdrawTransactionError, err
	}

	if err = wr.SaveWithdrawal(userID, orderNumber, sum); err != nil {
		_ = wr.GetDBStorage().Rollback()
		return repository.WithdrawTransactionError, err
	}

	if err = wr.GetDBStorage().Commit(); err != nil {
		return repository.WithdrawTransactionError, err
	}

	return 0, nil
}

func (ws *WithdrawService) Withdrawals(userID int) ([]interfaces.WithdrawInfo, error) {
	return ws.WithdrawRepository.Withdrawals(userID)
}
