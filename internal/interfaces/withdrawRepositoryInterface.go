package interfaces

import (
	"time"

	"github.com/shopspring/decimal"
)

// WithdrawRepositoryInterface определяет методы для работы с выводами средств.
// Этот интерфейс предоставляет операции для выполнения выводов и получения информации о выводах.
type WithdrawRepositoryInterface interface {
	// Withdraw выполняет вывод средств пользователя по номеру заказа и сумме.
	// Возвращает код результата выполнения и ошибку, если произошла ошибка.
	Withdraw(userID int, orderNumber string, sum decimal.Decimal) (int, error)

	// Withdrawals возвращает список выводов для указанного пользователя.
	// Возвращает список WithdrawInfo и ошибку, если произошла ошибка при запросе.
	Withdrawals(userID int) ([]WithdrawInfo, error)
}

// WithdrawInfo содержит информацию о выводе средств.
// Она включает номер заказа, сумму и дату обработки.
type WithdrawInfo struct {
	// OrderNumber представляет собой номер заказа.
	OrderNumber string `json:"order"`

	// Sum представляет собой сумму средств, которые были выведены.
	Sum float32 `json:"sum"`

	// ProcessedAt хранит время, когда вывод был обработан.
	ProcessedAt time.Time `json:"processed_at"`
}
