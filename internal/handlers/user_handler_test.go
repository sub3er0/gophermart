package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"gophermart/internal/interfaces"
	"gophermart/internal/middleware"
	"gophermart/internal/models"
	"gophermart/internal/repository"
	"gophermart/internal/service"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/shopspring/decimal"
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
	GetOrderIDFunc         func(orderNumber string, userID int) (int, error)
	SaveOrderFunc          func(orderNumber string, userID int) error
	UpdateOrderFunc        func(orderNumber string, accrual decimal.Decimal, status string) error
	GetUserOrdersFunc      func(userID int) ([]interfaces.OrderData, error)
}

func (os *MockOrderService) GetOrderID(orderNumber string, userID int) (int, error) {
	return os.GetOrderIDFunc(orderNumber, userID)
}

func (os *MockOrderService) SaveOrder(orderNumber string, userID int) error {
	return os.SaveOrderFunc(orderNumber, userID)
}

func (os *MockOrderService) UpdateOrder(orderNumber string, accrual decimal.Decimal, status string) error {
	return os.UpdateOrderFunc(orderNumber, accrual, status)
}

func (os *MockOrderService) GetUserOrders(userID int) ([]interfaces.OrderData, error) {
	return os.GetUserOrdersFunc(userID)
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

type MockUserBalanceRepository struct {
	CreateUserBalanceFunc func(models.User) error
	UpdateUserBalanceFunc func(accrual decimal.Decimal, userID int) error
	GetUserBalanceFunc    func(userID int) (interfaces.UserBalance, error)
}

func (ubr *MockUserBalanceRepository) UpdateUserBalance(accrual decimal.Decimal, userID int) error {
	return nil
}
func (ubr *MockUserBalanceRepository) GetUserBalance(userID int) (interfaces.UserBalance, error) {
	return ubr.GetUserBalanceFunc(userID)
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

func (m *MockTokenGenerator) _GenerateToken(user models.User) (string, error) {
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

func TestLogin_InvalidData(t *testing.T) {
	handler := UserHandler{}

	invalidJSON := `{username: "testuser"\", password: "password"}`
	body, _ := json.Marshal(invalidJSON)
	req := httptest.NewRequest("POST", "/api/user/login", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	http.HandlerFunc(handler.Login).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %v", rr.Code)
	}
}

func TestLogin_InvalidAuthenticate(t *testing.T) {
	mockUserService := &MockUserService{
		AuthenticateUserFunc: func(username, password string) (models.User, error) {
			return models.User{ID: 1, Username: username}, errors.New("error") // Успешная аутентификация
		},
	}
	handler := UserHandler{
		UserService: mockUserService,
	}

	creds := middleware.Credentials{Username: "testuser", Password: "password"}
	body, _ := json.Marshal(creds)
	req := httptest.NewRequest("POST", "/api/user/login", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	http.HandlerFunc(handler.Login).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %v", rr.Code)
	}
}

func TestLogin_InvalidGenerateToken(t *testing.T) {
	mockUserService := &MockUserService{
		AuthenticateUserFunc: func(username, password string) (models.User, error) {
			return models.User{ID: 1, Username: username}, nil // Успешная аутентификация
		},
	}
	mockTokenGen := &MockTokenGenerator{
		GenerateTokenFunc: func(user models.User) (string, error) {
			return "mockedToken", errors.New("error")
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

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %v", rr.Code)
	}
}

type MockNumberValidator struct {
	ValidateNumberFunc func() bool
}

func (mnv *MockNumberValidator) ValidateNumber(orderNumber string) bool {
	return mnv.ValidateNumberFunc()
}

type MockAccrualService struct {
	GetOrderInfoFunc func(accrualServerAddress string, orderNumber string) (service.RegisterResponse, error)
}

func (m *MockAccrualService) GetOrderInfo(accrualServerAddress string, orderNumber string) (service.RegisterResponse, error) {
	return m.GetOrderInfoFunc(accrualServerAddress, orderNumber)
}

func TestSaveOrder_Success(t *testing.T) {
	userID := 123
	orderNumber := 1234567890

	mockOrderService := &MockOrderService{
		GetOrderIDFunc: func(orderNumber string, userID int) (int, error) {
			return 0, nil
		},
		SaveOrderFunc: func(orderNumber string, userID int) error {
			return nil
		},
		UpdateOrderFunc: func(orderNumber string, accrual decimal.Decimal, status string) error {
			return nil
		},
	}

	mockUserBalanceService := &MockUserBalanceRepository{
		UpdateUserBalanceFunc: func(accrual decimal.Decimal, userID int) error {
			return nil
		},
	}

	mockDBStorage := &MockDBStorage{
		InitFunc:             func() error { return nil },
		BeginTransactionFunc: func() error { return nil },
		CommitFunc:           func() error { return nil },
		RollbackFunc:         func() error { return nil },
	}

	orderRepository := &MockOrderRepository{
		GetDBStorageFunc: func() interfaces.DBStorageInterface {
			return mockDBStorage
		},
	}
	mockOrderService.GetOrderRepositoryFunc = func() interfaces.OrderRepositoryInterface {
		return orderRepository
	}
	mockNumberValidator := &MockNumberValidator{
		ValidateNumberFunc: func() bool {
			return true
		},
	}
	mockAccrualService := &MockAccrualService{
		GetOrderInfoFunc: func(accrualServerAddress string, orderNumber string) (service.RegisterResponse, error) {
			return service.RegisterResponse{
				Order:   "1",
				Status:  "New",
				Accrual: decimal.NewFromFloat(1),
			}, nil
		},
	}
	userHandler := UserHandler{
		OrderService:       mockOrderService,
		UserBalanceService: mockUserBalanceService,
		DBConnectionString: "fake-connection-string",
		NumberValidator:    mockNumberValidator,
		AccrualService:     mockAccrualService,
	}

	body, _ := json.Marshal(orderNumber)
	req := httptest.NewRequest("POST", "/api/user/order", bytes.NewBuffer(body))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, userID))
	rr := httptest.NewRecorder()

	http.HandlerFunc(userHandler.SaveOrder).ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Errorf("Expected status 200, got %v", rr.Code)
	}
}

type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("forced read error")
}

func TestSaveOrder_InvalidData(t *testing.T) {
	userID := 123
	userHandler := UserHandler{}

	req, _ := http.NewRequest("POST", "/api/user/order", &errorReader{})
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, userID))
	rr := httptest.NewRecorder()

	http.HandlerFunc(userHandler.SaveOrder).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %v", rr.Code)
	}
}

func TestSaveOrder_InvalidDigit(t *testing.T) {
	userID := 123
	orderNumber := "1234567890asdf"

	userHandler := UserHandler{}

	body, _ := json.Marshal(orderNumber)
	req := httptest.NewRequest("POST", "/api/user/order", bytes.NewBuffer(body))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, userID))
	rr := httptest.NewRecorder()

	http.HandlerFunc(userHandler.SaveOrder).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %v", rr.Code)
	}
}

func TestSaveOrder_InvalidNumberValidator(t *testing.T) {
	userID := 123
	orderNumber := 1234567890

	mockNumberValidator := &MockNumberValidator{
		ValidateNumberFunc: func() bool {
			return false
		},
	}

	userHandler := UserHandler{
		NumberValidator: mockNumberValidator,
	}

	body, _ := json.Marshal(orderNumber)
	req := httptest.NewRequest("POST", "/api/user/order", bytes.NewBuffer(body))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, userID))
	rr := httptest.NewRecorder()

	http.HandlerFunc(userHandler.SaveOrder).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("Expected status 422, got %v", rr.Code)
	}
}

func TestSaveOrder_InvalidGetOrderID(t *testing.T) {
	userID := 123
	orderNumber := 1234567890

	mockOrderService := &MockOrderService{
		GetOrderIDFunc: func(orderNumber string, userID int) (int, error) {
			return 0, errors.New("error")
		},
		SaveOrderFunc: func(orderNumber string, userID int) error {
			return nil
		},
		UpdateOrderFunc: func(orderNumber string, accrual decimal.Decimal, status string) error {
			return nil
		},
	}
	mockNumberValidator := &MockNumberValidator{
		ValidateNumberFunc: func() bool {
			return true
		},
	}
	userHandler := UserHandler{
		OrderService:    mockOrderService,
		NumberValidator: mockNumberValidator,
	}

	body, _ := json.Marshal(orderNumber)
	req := httptest.NewRequest("POST", "/api/user/order", bytes.NewBuffer(body))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, userID))
	rr := httptest.NewRecorder()

	http.HandlerFunc(userHandler.SaveOrder).ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %v", rr.Code)
	}
}

func TestSaveOrder_LoadByAnotherUser(t *testing.T) {
	userID := 123
	orderNumber := 1234567890

	mockOrderService := &MockOrderService{
		GetOrderIDFunc: func(orderNumber string, userID int) (int, error) {
			return repository.OrderLoadedByAnotherUser, nil
		},
		SaveOrderFunc: func(orderNumber string, userID int) error {
			return nil
		},
		UpdateOrderFunc: func(orderNumber string, accrual decimal.Decimal, status string) error {
			return nil
		},
	}
	mockNumberValidator := &MockNumberValidator{
		ValidateNumberFunc: func() bool {
			return true
		},
	}
	userHandler := UserHandler{
		OrderService:    mockOrderService,
		NumberValidator: mockNumberValidator,
	}

	body, _ := json.Marshal(orderNumber)
	req := httptest.NewRequest("POST", "/api/user/order", bytes.NewBuffer(body))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, userID))
	rr := httptest.NewRecorder()

	http.HandlerFunc(userHandler.SaveOrder).ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("Expected status 409, got %v", rr.Code)
	}
}

func TestSaveOrder_LoadByThisUser(t *testing.T) {
	userID := 123
	orderNumber := 1234567890

	mockOrderService := &MockOrderService{
		GetOrderIDFunc: func(orderNumber string, userID int) (int, error) {
			return repository.OrderLoaderByThisUser, nil
		},
		SaveOrderFunc: func(orderNumber string, userID int) error {
			return nil
		},
		UpdateOrderFunc: func(orderNumber string, accrual decimal.Decimal, status string) error {
			return nil
		},
	}
	mockNumberValidator := &MockNumberValidator{
		ValidateNumberFunc: func() bool {
			return true
		},
	}
	userHandler := UserHandler{
		OrderService:    mockOrderService,
		NumberValidator: mockNumberValidator,
	}

	body, _ := json.Marshal(orderNumber)
	req := httptest.NewRequest("POST", "/api/user/order", bytes.NewBuffer(body))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, userID))
	rr := httptest.NewRecorder()

	http.HandlerFunc(userHandler.SaveOrder).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %v", rr.Code)
	}
}

func TestSaveOrder_SaveOrderError(t *testing.T) {
	userID := 123
	orderNumber := 1234567890

	mockOrderService := &MockOrderService{
		GetOrderIDFunc: func(orderNumber string, userID int) (int, error) {
			return -2, nil
		},
		SaveOrderFunc: func(orderNumber string, userID int) error {
			return errors.New("error")
		},
		UpdateOrderFunc: func(orderNumber string, accrual decimal.Decimal, status string) error {
			return nil
		},
	}
	mockNumberValidator := &MockNumberValidator{
		ValidateNumberFunc: func() bool {
			return true
		},
	}
	userHandler := UserHandler{
		OrderService:    mockOrderService,
		NumberValidator: mockNumberValidator,
	}

	body, _ := json.Marshal(orderNumber)
	req := httptest.NewRequest("POST", "/api/user/order", bytes.NewBuffer(body))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, userID))
	rr := httptest.NewRecorder()

	http.HandlerFunc(userHandler.SaveOrder).ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %v", rr.Code)
	}
}

func TestGetOrders_Success(t *testing.T) {
	mockOrders := []interfaces.OrderData{}

	mockOrderService := &MockOrderService{
		GetUserOrdersFunc: func(userID int) ([]interfaces.OrderData, error) {
			return mockOrders, nil
		},
	}

	userHandler := UserHandler{
		OrderService: mockOrderService,
	}

	req := httptest.NewRequest("GET", "/api/user/orders", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, 123)) // пользователь ID 123
	rr := httptest.NewRecorder()

	http.HandlerFunc(userHandler.GetOrders).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %v", rr.Code)
	}

	var response []interfaces.OrderData
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(response) != len(mockOrders) {
		t.Errorf("Expected %d orders, got %d", len(mockOrders), len(response))
	}
}

func TestGetBalance_Success(t *testing.T) {
	mockUserBalanceService := &MockUserBalanceRepository{
		GetUserBalanceFunc: func(userID int) (interfaces.UserBalance, error) {
			return interfaces.UserBalance{
				Current:   100,
				Withdrawn: 100,
			}, nil
		},
	}

	userHandler := UserHandler{
		UserBalanceService: mockUserBalanceService,
	}

	req := httptest.NewRequest("GET", "/api/user/balance", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, 123)) // Устанавливаем ID пользователя
	rr := httptest.NewRecorder()

	http.HandlerFunc(userHandler.GetBalance).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %v", rr.Code)
	}

	var response interfaces.UserBalance
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Withdrawn != 100.00 {
		t.Errorf("Expected withdrawn 100.00, got %.2f", response.Withdrawn)
	}

	if response.Current != 100.00 {
		t.Errorf("Expected current 100.00, got %.2f", response.Current)
	}
}

type MockWithdrawRepository struct {
	WithdrawFunc    func() (int, error)
	WithdrawalsFunc func() ([]interfaces.WithdrawInfo, error)
}

func (mwr *MockWithdrawRepository) Withdraw(userID int, orderNumber string, sum decimal.Decimal) (int, error) {
	return mwr.WithdrawFunc()
}
func (mwr *MockWithdrawRepository) Withdrawals(userID int) ([]interfaces.WithdrawInfo, error) {
	return mwr.WithdrawalsFunc()
}

func TestWithdraw_Success(t *testing.T) {
	mockOrderService := &MockOrderService{
		GetOrderIDFunc: func(orderNumber string, userID int) (int, error) {
			return 0, nil
		},
	}

	mockWithdrawService := &MockWithdrawRepository{
		WithdrawFunc: func() (int, error) {
			return 0, nil
		},
	}

	userHandler := UserHandler{
		OrderService:    mockOrderService,
		WithdrawService: mockWithdrawService,
	}

	withdrawData := models.Withdraw{
		Order: "123456",
		Sum:   decimal.NewFromFloat(100.0),
	}
	body, _ := json.Marshal(withdrawData)
	req := httptest.NewRequest("POST", "/api/user/withdraw", bytes.NewBuffer(body))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, 123)) // Установка ID пользователя
	rr := httptest.NewRecorder()

	http.HandlerFunc(userHandler.Withdraw).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %v", rr.Code)
	}
}

func TestWithdrawals_Success(t *testing.T) {
	mockWithdrawService := &MockWithdrawRepository{
		WithdrawalsFunc: func() ([]interfaces.WithdrawInfo, error) {
			return []interfaces.WithdrawInfo{
				{OrderNumber: "123456", Sum: 100.0},
				{OrderNumber: "789012", Sum: 50.0},
			}, nil
		},
	}

	userHandler := UserHandler{
		WithdrawService: mockWithdrawService,
	}

	req := httptest.NewRequest("GET", "/api/user/withdrawals", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, 123)) // Устанавливаем ID пользователя
	rr := httptest.NewRecorder()

	http.HandlerFunc(userHandler.Withdrawals).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %v", rr.Code)
	}

	var response []models.Withdraw
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(response) != 2 {
		t.Errorf("Expected 2 withdrawals, got %d", len(response))
	}
}
