package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v4"
	"github.com/shopspring/decimal"
	"gophermart/internal/interfaces"
	"gophermart/internal/middleware"
	"gophermart/internal/models"
	"gophermart/internal/repository"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockUserService struct {
	RegisterUserFunc      func(models.User) (models.User, error)
	AuthenticateUserFunc  func(string, string) (models.User, error)
	GetUserRepositoryFunc func() interfaces.UserRepositoryInterface
}

func (m *MockUserService) GetUserID(username string) int {
	return 1
}

func (m *MockUserService) RegisterUser(u models.User) (models.User, error) {
	return m.RegisterUserFunc(u)
}

func (m *MockUserService) AuthenticateUser(username, password string) (models.User, error) {
	return m.AuthenticateUserFunc(username, password)
}

func (m *MockUserService) GetUserRepository() interfaces.UserRepositoryInterface {
	return m.GetUserRepositoryFunc()
}

type MockUserRepository struct {
	GetUserIDFunc func(string) int
}

func (ur *MockUserRepository) CreateUser(user models.User) (int, error) {
	return -1, nil
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

type MockOrderService struct {
	RegisterUserFunc       func(models.User) (models.User, error)
	AuthenticateUserFunc   func(string, string) (models.User, error)
	GetOrderRepositoryFunc func() interfaces.OrderRepositoryInterface
}

func (os *MockOrderService) GetOrderID(orderNumber string, userID int) (int, error) {
	return -1, nil
}

func (os *MockOrderService) SaveOrder(orderNumber string, userID int) error {
	return nil
}

func (os *MockOrderService) UpdateOrder(orderNumber string, accrual decimal.Decimal, status string) error {
	return nil
}

func (os *MockOrderService) GetUserOrders(userID int) ([]interfaces.OrderData, error) {
	return []interfaces.OrderData{}, nil
}

func (os *MockOrderService) GetOrderRepository() interfaces.OrderRepositoryInterface {
	return os.GetOrderRepositoryFunc()
}

type MockOrderRepository struct {
	GetDBStorageFunc func() interfaces.DBStorageInterface
}

func (or *MockOrderRepository) GetOrderID(orderNumber string, userID int) (int, error) {
	return -1, nil
}

func (or *MockOrderRepository) SaveOrder(orderNumber string, userID int) error {
	return nil
}

func (or *MockOrderRepository) UpdateOrder(orderNumber string, accrual decimal.Decimal, status string) error {
	return nil
}

func (or *MockOrderRepository) GetUserOrders(userID int) ([]interfaces.OrderData, error) {
	return []interfaces.OrderData{}, nil
}

func (or *MockOrderRepository) GetDBStorage() interfaces.DBStorageInterface {
	return or.GetDBStorageFunc()
}

type MockDBStorage struct {
	InitFunc             func() error
	BeginTransactionFunc func() error
	CommitFunc           func() error
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

type MockUserBalanceRepository struct {
	CreateUserBalanceFunc func(models.User) error
}

func (ubr *MockUserBalanceRepository) UpdateUserBalance(accrual decimal.Decimal, userID int) error {
	return nil
}
func (ubr *MockUserBalanceRepository) GetUserBalance(userID int) (interfaces.UserBalance, error) {
	return interfaces.UserBalance{}, nil
}
func (ubr *MockUserBalanceRepository) CreateUserBalance(user models.User) error {
	return ubr.CreateUserBalanceFunc(user)
}

func TestRegister(t *testing.T) {
	mockUserService := &MockUserService{
		RegisterUserFunc: func(user models.User) (models.User, error) {
			return user, nil
		},
		GetUserRepositoryFunc: func() interfaces.UserRepositoryInterface {
			return &MockUserRepository{
				GetUserIDFunc: func(username string) int {
					return repository.UserNotFound
				},
			}
		},
	}
	dbStorage := &MockDBStorage{
		InitFunc: func() error {
			return nil
		},
		BeginTransactionFunc: func() error {
			return nil
		},
		CommitFunc: func() error {
			return nil
		},
	}
	orderService := &MockOrderService{
		GetOrderRepositoryFunc: func() interfaces.OrderRepositoryInterface {
			return &MockOrderRepository{
				GetDBStorageFunc: func() interfaces.DBStorageInterface {
					return dbStorage
				},
			}
		},
	}
	userBalanceService := &MockUserBalanceRepository{
		CreateUserBalanceFunc: func(user models.User) error {
			return nil
		},
	}
	mockTokenGen := &MockTokenGenerator{
		GenerateTokenFunc: func(user models.User) (string, error) {
			return "mockedToken", nil
		},
	}
	handler := UserHandler{
		UserService:        mockUserService,
		OrderService:       orderService,
		UserBalanceService: userBalanceService,
		DBConnectionString: "fake-connection-string",
		TokenGenerator:     mockTokenGen,
	}

	user := models.User{Username: "testuser", Password: "password"}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/api/user/register", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	http.HandlerFunc(handler.Register).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %v", rr.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	if message, exists := response["message"]; !exists || message != "User registered successfully" {
		t.Errorf("Expected success message, got %v", message)
	}
}

func TestRegister_InvalidData(t *testing.T) {
	user := "{ invalid json "
	body := bytes.NewBuffer([]byte(user))
	req := httptest.NewRequest("POST", "/api/user/register", body)
	rr := httptest.NewRecorder()

	var handler UserHandler
	http.HandlerFunc(handler.Register).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid data, got %v", rr.Code)
	}
}

func TestRegister_DatabaseError(t *testing.T) {
	mockUserService := &MockUserService{
		RegisterUserFunc: func(user models.User) (models.User, error) {
			return user, nil
		},
		GetUserRepositoryFunc: func() interfaces.UserRepositoryInterface {
			return &MockUserRepository{
				GetUserIDFunc: func(username string) int {
					return -2
				},
			}
		},
	}
	handler := UserHandler{
		UserService:        mockUserService,
		OrderService:       nil,
		UserBalanceService: nil,
		DBConnectionString: "fake-connection-string",
	}

	user := models.User{Username: "testuser", Password: "password"}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/api/user/register", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	http.HandlerFunc(handler.Register).ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %v", rr.Code)
	}
}

func TestRegister_RegisterError(t *testing.T) {
	mockUserService := &MockUserService{
		RegisterUserFunc: func(user models.User) (models.User, error) {
			return user, nil
		},
		GetUserRepositoryFunc: func() interfaces.UserRepositoryInterface {
			return &MockUserRepository{
				GetUserIDFunc: func(username string) int {
					return 1
				},
			}
		},
	}
	handler := UserHandler{
		UserService:        mockUserService,
		OrderService:       nil,
		UserBalanceService: nil,
		DBConnectionString: "fake-connection-string",
	}

	user := models.User{Username: "testuser", Password: "password"}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/api/user/register", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	http.HandlerFunc(handler.Register).ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("Expected status 409, got %v", rr.Code)
	}
}

func TestRegister_InitError(t *testing.T) {
	mockUserService := &MockUserService{
		RegisterUserFunc: func(user models.User) (models.User, error) {
			return user, nil
		},
		GetUserRepositoryFunc: func() interfaces.UserRepositoryInterface {
			return &MockUserRepository{
				GetUserIDFunc: func(username string) int {
					return repository.UserNotFound
				},
			}
		},
	}
	dbStorage := &MockDBStorage{
		InitFunc: func() error {
			return errors.New("error")
		},
		BeginTransactionFunc: func() error {
			return nil
		},
	}
	orderService := &MockOrderService{
		GetOrderRepositoryFunc: func() interfaces.OrderRepositoryInterface {
			return &MockOrderRepository{
				GetDBStorageFunc: func() interfaces.DBStorageInterface {
					return dbStorage
				},
			}
		},
	}
	handler := UserHandler{
		UserService:        mockUserService,
		OrderService:       orderService,
		UserBalanceService: nil,
		DBConnectionString: "fake-connection-string",
	}

	user := models.User{Username: "testuser", Password: "password"}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/api/user/register", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	http.HandlerFunc(handler.Register).ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %v", rr.Code)
	}
}

func TestRegister_BeginTransactionError(t *testing.T) {
	mockUserService := &MockUserService{
		RegisterUserFunc: func(user models.User) (models.User, error) {
			return user, nil
		},
		GetUserRepositoryFunc: func() interfaces.UserRepositoryInterface {
			return &MockUserRepository{
				GetUserIDFunc: func(username string) int {
					return repository.UserNotFound
				},
			}
		},
	}
	dbStorage := &MockDBStorage{
		InitFunc: func() error {
			return nil
		},
		BeginTransactionFunc: func() error {
			return errors.New("error")
		},
	}
	orderService := &MockOrderService{
		GetOrderRepositoryFunc: func() interfaces.OrderRepositoryInterface {
			return &MockOrderRepository{
				GetDBStorageFunc: func() interfaces.DBStorageInterface {
					return dbStorage
				},
			}
		},
	}
	handler := UserHandler{
		UserService:        mockUserService,
		OrderService:       orderService,
		UserBalanceService: nil,
		DBConnectionString: "fake-connection-string",
	}

	user := models.User{Username: "testuser", Password: "password"}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/api/user/register", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	http.HandlerFunc(handler.Register).ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %v", rr.Code)
	}
}

func TestRegister_RegisterUserError(t *testing.T) {
	mockUserService := &MockUserService{
		RegisterUserFunc: func(user models.User) (models.User, error) {
			return user, errors.New("error")
		},
		GetUserRepositoryFunc: func() interfaces.UserRepositoryInterface {
			return &MockUserRepository{
				GetUserIDFunc: func(username string) int {
					return repository.UserNotFound
				},
			}
		},
	}
	dbStorage := &MockDBStorage{
		InitFunc: func() error {
			return nil
		},
		BeginTransactionFunc: func() error {
			return nil
		},
	}
	orderService := &MockOrderService{
		GetOrderRepositoryFunc: func() interfaces.OrderRepositoryInterface {
			return &MockOrderRepository{
				GetDBStorageFunc: func() interfaces.DBStorageInterface {
					return dbStorage
				},
			}
		},
	}
	handler := UserHandler{
		UserService:        mockUserService,
		OrderService:       orderService,
		UserBalanceService: nil,
		DBConnectionString: "fake-connection-string",
	}

	user := models.User{Username: "testuser", Password: "password"}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/api/user/register", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	http.HandlerFunc(handler.Register).ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %v", rr.Code)
	}
}

func TestRegister_CreateUserBalanceError(t *testing.T) {
	mockUserService := &MockUserService{
		RegisterUserFunc: func(user models.User) (models.User, error) {
			return user, nil
		},
		GetUserRepositoryFunc: func() interfaces.UserRepositoryInterface {
			return &MockUserRepository{
				GetUserIDFunc: func(username string) int {
					return repository.UserNotFound
				},
			}
		},
	}
	userBalanceService := &MockUserBalanceRepository{
		CreateUserBalanceFunc: func(user models.User) error {
			return errors.New("error")
		},
	}
	dbStorage := &MockDBStorage{
		InitFunc: func() error {
			return nil
		},
		BeginTransactionFunc: func() error {
			return nil
		},
	}
	orderService := &MockOrderService{
		GetOrderRepositoryFunc: func() interfaces.OrderRepositoryInterface {
			return &MockOrderRepository{
				GetDBStorageFunc: func() interfaces.DBStorageInterface {
					return dbStorage
				},
			}
		},
	}
	handler := UserHandler{
		UserService:        mockUserService,
		OrderService:       orderService,
		UserBalanceService: userBalanceService,
		DBConnectionString: "fake-connection-string",
	}

	user := models.User{Username: "testuser", Password: "password"}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/api/user/register", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	http.HandlerFunc(handler.Register).ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %v", rr.Code)
	}
}

func TestRegister_CommitError(t *testing.T) {
	mockUserService := &MockUserService{
		RegisterUserFunc: func(user models.User) (models.User, error) {
			return user, nil
		},
		GetUserRepositoryFunc: func() interfaces.UserRepositoryInterface {
			return &MockUserRepository{
				GetUserIDFunc: func(username string) int {
					return repository.UserNotFound
				},
			}
		},
	}
	userBalanceService := &MockUserBalanceRepository{
		CreateUserBalanceFunc: func(user models.User) error {
			return nil
		},
	}
	dbStorage := &MockDBStorage{
		InitFunc: func() error {
			return nil
		},
		BeginTransactionFunc: func() error {
			return nil
		},
		CommitFunc: func() error {
			return errors.New("error")
		},
	}
	orderService := &MockOrderService{
		GetOrderRepositoryFunc: func() interfaces.OrderRepositoryInterface {
			return &MockOrderRepository{
				GetDBStorageFunc: func() interfaces.DBStorageInterface {
					return dbStorage
				},
			}
		},
	}
	handler := UserHandler{
		UserService:        mockUserService,
		OrderService:       orderService,
		UserBalanceService: userBalanceService,
		DBConnectionString: "fake-connection-string",
	}

	user := models.User{Username: "testuser", Password: "password"}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/api/user/register", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	http.HandlerFunc(handler.Register).ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %v", rr.Code)
	}
}

type MockTokenGenerator struct {
	GenerateTokenFunc func(models.User) (string, error)
}

func (m *MockTokenGenerator) GenerateToken(user models.User) (string, error) {
	return m.GenerateTokenFunc(user)
}

func TestRegister_GenerateToken(t *testing.T) {
	mockUserService := &MockUserService{
		RegisterUserFunc: func(user models.User) (models.User, error) {
			return user, nil
		},
		GetUserRepositoryFunc: func() interfaces.UserRepositoryInterface {
			return &MockUserRepository{
				GetUserIDFunc: func(username string) int {
					return repository.UserNotFound
				},
			}
		},
	}
	userBalanceService := &MockUserBalanceRepository{
		CreateUserBalanceFunc: func(user models.User) error {
			return nil
		},
	}
	mockTokenGen := &MockTokenGenerator{
		GenerateTokenFunc: func(user models.User) (string, error) {
			return "mockedToken", errors.New("error")
		},
	}
	dbStorage := &MockDBStorage{
		InitFunc: func() error {
			return nil
		},
		BeginTransactionFunc: func() error {
			return nil
		},
		CommitFunc: func() error {
			return nil
		},
	}
	orderService := &MockOrderService{
		GetOrderRepositoryFunc: func() interfaces.OrderRepositoryInterface {
			return &MockOrderRepository{
				GetDBStorageFunc: func() interfaces.DBStorageInterface {
					return dbStorage
				},
			}
		},
	}
	handler := UserHandler{
		UserService:        mockUserService,
		OrderService:       orderService,
		UserBalanceService: userBalanceService,
		DBConnectionString: "fake-connection-string",
		TokenGenerator:     mockTokenGen,
	}

	user := models.User{Username: "testuser", Password: "password"}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/api/user/register", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	http.HandlerFunc(handler.Register).ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %v", rr.Code)
	}
}

func TestLogin(t *testing.T) {
	mockUserService := &MockUserService{
		AuthenticateUserFunc: func(username, password string) (models.User, error) {
			return models.User{ID: 1, Username: username}, nil // Успешная аутентификация
		},
	}
	mockTokenGen := &MockTokenGenerator{
		GenerateTokenFunc: func(user models.User) (string, error) {
			return "mockedToken", nil
		},
	}
	handler := UserHandler{
		UserService:    mockUserService,
		TokenGenerator: mockTokenGen,
	}

	creds := middleware.Credentials{Username: "testuser", Password: "password"}
	body, _ := json.Marshal(creds)
	req := httptest.NewRequest("POST", "/api/user/login", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	http.HandlerFunc(handler.Login).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %v", rr.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	if message, exists := response["message"]; !exists || message != "Login successful" {
		t.Errorf("Expected success message, got %v", message)
	}
}
