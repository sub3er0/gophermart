package interfaces

import (
	"time"
)

type OrderRepositoryInterface interface {
	IsOrderExist(orderNumber string, userID int) (int, error)
	SaveOrder(orderNumber string, userID int) error
	UpdateOrder(orderNumber string, accrual float32, status string) error
	GetUserOrders(userID int) ([]OrderData, error)
	GetDBStorage() DBStorageInterface
}

type OrderData struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float32   `json:"accrual"`
	UploadedAt time.Time `json:"uploaded_at"`
}
