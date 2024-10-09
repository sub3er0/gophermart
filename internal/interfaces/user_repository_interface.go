package interfaces

import (
	"gophermart/internal/models"
)

type UserRepositoryInterface interface {
	CreateUser(user models.User) (int, error)
	IsUserExists(username string) int
	GetUserByUsername(username string) (models.User, error)
	GetDBStorage() DBStorageInterface
}
