package interfaces

import (
	"time"

	"github.com/shopspring/decimal"
)

// OrderRepositoryInterface определяет методы для работы с заказами.
// Этот интерфейс позволяет взаимодействовать с хранилищем данных заказов,
// предоставляя операции для получения, сохранения и обновления информации о заказах.
type OrderRepositoryInterface interface {
	// GetOrderID возвращает ID заказа по номеру заказа и ID пользователя.
	// Возвращает ошибку, если заказ не найден или произошла ошибка при запросе.
	GetOrderID(orderNumber string, userID int) (int, error)

	// SaveOrder сохраняет новый заказ в хранилище данных.
	// Возвращает ошибку, если произошла ошибка при сохранении.
	SaveOrder(orderNumber string, userID int) error

	// UpdateOrder обновляет информацию о заказе в хранилище данных,
	// включая его номер, начисление и статус.
	// Возвращает ошибку, если произошла ошибка при обновлении.
	UpdateOrder(orderNumber string, accrual decimal.Decimal, status string) error

	// GetUserOrders возвращает список заказов для указанного пользователя по его ID.
	// Возвращает список объектов OrderData и ошибку, если произошла ошибка при запросе.
	GetUserOrders(userID int) ([]OrderData, error)

	// GetDBStorage возвращает объект для доступа к хранилищу данных.
	GetDBStorage() DBStorageInterface
}

// OrderData представляет собой данные заказа.
// Она содержит информацию о номере заказа, его статусе,
// начислении и времени загрузки.
type OrderData struct {
	// Number является номером заказа.
	Number string `json:"number"`

	// Status указывает текущий статус заказа.
	Status string `json:"status"`

	// Accrual представляет собой сумму начисления, связанного с заказом.
	Accrual float32 `json:"accrual"`

	// UploadedAt хранит время загрузки заказа.
	UploadedAt time.Time `json:"uploaded_at"`
}
