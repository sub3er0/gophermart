package interfaces

import (
	"gophermart/internal/models"
)

type UserServiceInterface interface {
	GetUserID(username string) int
	CreateUser(user models.User) (models.User, error)
	AuthenticateUser(username, password string) (models.User, error)
	GetUserRepository() UserRepositoryInterface
	RegisterUser(user models.User, userBalanceRepository UserBalanceRepositoryInterface) error
}
