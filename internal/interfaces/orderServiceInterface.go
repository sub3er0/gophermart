package interfaces

import "github.com/shopspring/decimal"

// OrderServiceInterface определяет методы для работы с заказами.
// Этот интерфейс предоставляет операции для получения, сохранения и обновления информации о заказах.
type OrderServiceInterface interface {
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

	// GetOrderRepository возвращает интерфейс для доступа к репозиторию заказов.
	// Это может быть полезно для получения методов работы с базой данных для заказов.
	GetOrderRepository() OrderRepositoryInterface
}
