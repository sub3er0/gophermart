package service

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"gophermart/internal/interfaces"
	"gophermart/internal/models"
	"gophermart/internal/repository"
	"log"
)

var (
	ErrFailedToRegister = errors.New("failed to register user")
	ErrDb               = errors.New("internal database error")
)

type UserService struct {
	UserRepository   *repository.UserRepository
	connectionString string
}

func (us *UserService) SetConnectionString(connectionString string) {
	us.connectionString = connectionString
}

func (us *UserService) GetUserID(username string) int {
	return us.UserRepository.GetUserID(username)
}

func (us *UserService) CreateUser(user models.User) (models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

	if err != nil {
		return user, err
	}

	user.Password = string(hashedPassword)

	user.ID, err = us.UserRepository.CreateUser(user)

	if err != nil {
		return user, err
	}

	return user, nil
}

func (us *UserService) AuthenticateUser(username, password string) (models.User, error) {
	user, err := us.UserRepository.GetUserByUsername(username)

	if err != nil {
		return models.User{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (us *UserService) GetUserRepository() interfaces.UserRepositoryInterface {
	return us.UserRepository
}

func (us *UserService) RegisterUser(user models.User, userBalanceRepository interfaces.UserBalanceRepositoryInterface) error {
	userRepository := us.GetUserRepository()
	dbStorage := userRepository.GetDBStorage()
	err := dbStorage.Init(us.connectionString)

	if err != nil {
		log.Fatalf("Error while initializing db connection: %v", err)
	}

	err = dbStorage.BeginTransaction()

	if err != nil {
		return ErrDb
	}

	if user, err = us.CreateUser(user); err != nil {
		_ = dbStorage.Rollback()
		return ErrFailedToRegister
	}

	if err = userBalanceRepository.CreateUserBalance(user); err != nil {
		_ = dbStorage.Rollback()
		return ErrFailedToRegister
	}

	if err := dbStorage.Commit(); err != nil {
		return ErrDb
	}

	dbStorage.Close()

	return nil
}
