package interfaces

import (
	"gophermart/internal/models"
)

// UserRepositoryInterface определяет методы для работы с пользователями.
// Этот интерфейс предоставляет операции для создания пользователей,
// получения их ID и извлечения информации о пользователе.
type UserRepositoryInterface interface {
	// CreateUser создает нового пользователя и возвращает его уникальный ID.
	// Возвращает ошибку, если произошла ошибка при создании.
	CreateUser(user models.User) (int, error)

	// GetUserID возвращает ID пользователя по его имени пользователя.
	// Возвращает ноль, если пользователь не найден.
	GetUserID(username string) int

	// GetUserByUsername извлекает информацию о пользователе по его имени пользователя.
	// Возвращает объект User и ошибку, если произошла ошибка при запросе.
	GetUserByUsername(username string) (models.User, error)

	// GetDBStorage возвращает интерфейс для доступа к хранилищу данных.
	GetDBStorage() DBStorageInterface
}
