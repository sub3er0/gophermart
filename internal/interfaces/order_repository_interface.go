package interfaces

import "gophermart/internal/repository"

type OrderRepositoryInterface interface {
	IsOrderExist(orderNumber string, userID int) (int, error)
	SaveOrder(orderNumber string, userID int) error
	UpdateOrder(orderNumber string, accrual float32, status string) error
	GetUserOrders(userID int) ([]repository.OrderData, error)
}
