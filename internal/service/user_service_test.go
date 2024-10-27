package service

import (
	"database/sql"
	"github.com/jackc/pgx/v4"
	"gophermart/internal/interfaces"
	"gophermart/internal/models"
	"testing"
)

type MockDBStorage struct {
	InitFunc             func() error
	BeginTransactionFunc func() error
	CommitFunc           func() error
	RollbackFunc         func() error
}

func (dbs *MockDBStorage) Init(connectionString string) error {
	return dbs.InitFunc()
}

func (dbs *MockDBStorage) Exec(query string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}
func (dbs *MockDBStorage) Select(query string, args ...interface{}) (pgx.Rows, error) {
	return nil, nil
}
func (dbs *MockDBStorage) Close()                                             {}
func (dbs *MockDBStorage) BeginTransaction() error                            { return dbs.BeginTransactionFunc() }
func (dbs *MockDBStorage) Rollback() error                                    { return nil }
func (dbs *MockDBStorage) Commit() error                                      { return dbs.CommitFunc() }
func (dbs *MockDBStorage) QueryRow(query string, args ...interface{}) pgx.Row { return nil }

type MockUserRepository struct {
	GetUserIDFunc  func(string) int
	CreateUserFunc func(models.User) (int, error)
}

func (ur *MockUserRepository) CreateUser(user models.User) (int, error) {
	return ur.CreateUserFunc(user)
}

func (ur *MockUserRepository) GetUserID(username string) int {
	return ur.GetUserIDFunc(username)
}

func (ur *MockUserRepository) GetUserByUsername(username string) (models.User, error) {
	return models.User{}, nil
}

func (ur *MockUserRepository) GetDBStorage() interfaces.DBStorageInterface {
	return &MockDBStorage{}
}

func TestRegisterUser_Success(t *testing.T) {
	user := models.User{
		Username: "username",
		Password: "password",
	}
	userRepository := MockUserRepository{
		CreateUserFunc: func(user models.User) (int, error) {
			return 1, nil
		},
	}

	userService := &UserService{
		UserRepository: &userRepository,
	}

	_, err := userService.RegisterUser(user)

	if err != nil {
		t.Errorf("Expected err == nil")
	}
}
