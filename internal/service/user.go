package service

import (
	"gophermart/internal/interfaces"
	"gophermart/internal/models"

	"golang.org/x/crypto/bcrypt"
)

type UserRepositoryInterface interface {
	GetDBStorage() interfaces.DBStorageInterface
	CreateUser(user models.User) (int, error)
	GetUserID(username string) int
	GetUserByUsername(username string) (models.User, error)
}

type UserService struct {
	UserRepository UserRepositoryInterface
}

func (us *UserService) GetUserID(username string) int {
	return us.UserRepository.GetUserID(username)
}

func (us *UserService) RegisterUser(user models.User) (models.User, error) {
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
