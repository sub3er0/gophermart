package interfaces

import (
	"gophermart/internal/models"

	"github.com/shopspring/decimal"
)

// UserBalanceRepositoryInterface определяет методы для работы с балансом пользователя.
// Этот интерфейс предоставляет операции для обновления, получения и создания баланса пользователя.
type UserBalanceRepositoryInterface interface {
	// UpdateUserBalance обновляет баланс пользователя на основе начислений.
	// Возвращает ошибку, если произошла ошибка при обновлении.
	UpdateUserBalance(accrual decimal.Decimal, userID int) error

	// GetUserBalance возвращает текущий баланс пользователя по его ID.
	// Возвращает объект UserBalance и ошибку, если произошла ошибка при запросе.
	GetUserBalance(userID int) (UserBalance, error)

	// CreateUserBalance создает новый баланс для указанного пользователя.
	// Возвращает ошибку, если произошла ошибка при создании.
	CreateUserBalance(user models.User) error
}

// UserBalance представляет собой данные о балансе пользователя.
// Она содержит информацию о текущем балансе и снятых средствах.
type UserBalance struct {
	// Current представляет собой текущий баланс пользователя.
	Current float32 `json:"current"`

	// Withdrawn указывает на сумму средств, которые были сняты пользователем.
	Withdrawn float32 `json:"withdrawn"`
}
