package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"gophermart/internal/interfaces"
	"gophermart/internal/middleware"
	"gophermart/internal/models"
	"gophermart/internal/repository"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockUserService struct {
	Repo MockUserRepository
}

type MockOrderService struct {
	Repo MockUserRepository
}
type MockWithdrawService struct{}
type MockUserBalanceService struct{}

type MockUserRepository struct{}

func (m *MockUserRepository) IsUserExists(username string) int {
	if username == "existingUser" {
		return 1
	}
	return -1
}

func (m *MockUserRepository) CreateUser(user models.User) (int, error) {
	return 1, nil
}

func (m *MockUserRepository) GetUserByUsername(username string) (models.User, error) {
	if username == "validUser" {
		return models.User{Username: "validUser"}, nil
	}
	return models.User{}, nil
}

func (m *MockUserService) GetUserRepository() interfaces.UserRepositoryInterface {
	return &m.Repo
}

func (m *MockUserService) RegisterUser(user models.User) (models.User, error) {
	return models.User{
		ID:       0,
		Username: "",
		Password: "",
	}, nil
}

func (m *MockUserService) AuthenticateUser(username, password string) (models.User, error) {
	if username == "validUser" && password == "validPassword" {
		return models.User{Username: "validUser"}, nil
	}
	return models.User{}, nil
}

func (m *MockUserService) IsUserExist(username string) int {
	if username == "existingUser" {
		return 1 // user exists
	}
	return -1
}

//type OrderServiceInterface interface {
//	IsOrderExist(orderNumber string, userID int) (int, error)
//	SaveOrder(orderNumber string, userID int) error
//	UpdateOrder(orderNumber string, accrual float32, status string) error
//	GetUserOrders(userID int) ([]repository.OrderData, error)
//	GetOrderRepository() interfaces.OrderRepositoryInterface
//}

func (mos *MockOrderService) IsOrderExist(orderNumber string, userID int) (int, error) {
	return 1, nil
}

func (mos *MockOrderService) SaveOrder(orderNumber string, userID int) error {
	return nil
}

func (mos *MockOrderService) UpdateOrder(orderNumber string, accrual float32, status string) error {
	return nil
}

func (mos *MockOrderService) GetUserOrders(userID int) ([]repository.OrderData, error) {
	orderData := []repository.OrderData{}
	return orderData, nil
}

func (mos *MockOrderService) GetOrderRepository() *repository.OrderRepository {
	return mos.Repo
}

func TestRegister(t *testing.T) {
	tests := []struct {
		name            string
		input           models.User
		expectedCode    int
		expectedMessage string
	}{
		{
			name: "successful registration",
			input: models.User{
				Username: "newUser",
				Password: "password123",
			},
			expectedCode:    http.StatusOK,
			expectedMessage: "User registered successfully",
		},
		{
			name: "user already exists",
			input: models.User{
				Username: "existingUser",
				Password: "password123",
			},
			expectedCode:    http.StatusConflict,
			expectedMessage: "Failed to register user",
		},
		{
			name: "invalid input",
			input: models.User{
				Username: "",
				Password: "password123",
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockUserRepository{}
			userService := &MockUserService{Repo: *mockRepo}
			userHandler := &UserHandler{
				UserService: userService,
			}

			body, _ := json.Marshal(tt.input)
			req, err := http.NewRequest("POST", "/register", bytes.NewReader(body))
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(userHandler.Register)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)

			if tt.expectedMessage != "" {
				var response map[string]string
				json.NewDecoder(rr.Body).Decode(&response)
				assert.Equal(t, tt.expectedMessage, response["message"])
			}
		})
	}
}

func TestLogin(t *testing.T) {
	tests := []struct {
		name            string
		input           middleware.Credentials
		expectedCode    int
		expectedMessage string
	}{
		{
			name: "successful login",
			input: middleware.Credentials{
				Username: "validUser",
				Password: "validPassword",
			},
			expectedCode:    http.StatusOK,
			expectedMessage: "Login successful",
		},
		{
			name: "unauthorized login",
			input: middleware.Credentials{
				Username: "invalidUser",
				Password: "invalidPassword",
			},
			expectedCode:    http.StatusUnauthorized,
			expectedMessage: "Unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userHandler := &UserHandler{
				UserService: &MockUserService{},
			}

			body, _ := json.Marshal(tt.input)
			req, err := http.NewRequest("POST", "/login", bytes.NewReader(body))
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(userHandler.Login)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)

			if tt.expectedMessage != "" {
				var response map[string]string
				json.NewDecoder(rr.Body).Decode(&response)
				assert.Equal(t, tt.expectedMessage, response["message"])
			}
		})
	}
}
