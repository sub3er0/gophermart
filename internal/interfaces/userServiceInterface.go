package interfaces

import (
	"gophermart/internal/models"
)

// UserServiceInterface определяет методы для работы с пользователями.
// Этот интерфейс предоставляет операции для получения ID пользователя,
// регистрации и аутентификации пользователей.
type UserServiceInterface interface {
	// GetUserID возвращает ID пользователя по его имени пользователя.
	// Возвращает 0, если пользователь не найден.
	GetUserID(username string) int

	// RegisterUser регистрирует нового пользователя и возвращает его данные.
	// Возвращает ошибку, если произошла ошибка при регистрации.
	RegisterUser(user models.User) (models.User, error)

	// AuthenticateUser аутентифицирует пользователя по имени пользователя и паролю.
	// Возвращает объект User и ошибку, если учетные данные неверны или произошла ошибка.
	AuthenticateUser(username, password string) (models.User, error)

	// GetUserRepository возвращает интерфейс для доступа к репозиторию пользователей.
	GetUserRepository() UserRepositoryInterface
}
