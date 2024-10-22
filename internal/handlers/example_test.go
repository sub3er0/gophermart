package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"gophermart/internal/interfaces"
	"gophermart/internal/middleware"
	"gophermart/internal/models"
	"net/http"
	"net/http/httptest"
)

func ExampleUserHandler_Register() {
	mockUserService := &MockUserService{
		RegisterUserFunc: func(user models.User) (models.User, error) {
			return user, nil
		},
	}

	userHandler := UserHandler{
		UserService: mockUserService,
	}

	user := models.User{Username: "testuser", Password: "password"}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/api/user/register", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	http.HandlerFunc(userHandler.Register).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		return
	}
}

func ExampleUserHandler_Login() {
	mockUserService := &MockUserService{
		AuthenticateUserFunc: func(username, password string) (models.User, error) {
			return models.User{ID: 1, Username: username}, nil
		},
	}

	userHandler := UserHandler{
		UserService: mockUserService,
	}

	creds := middleware.Credentials{Username: "testuser", Password: "password"}
	body, _ := json.Marshal(creds)
	req := httptest.NewRequest("POST", "/api/user/login", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	http.HandlerFunc(userHandler.Login).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		return
	}
}

func ExampleUserHandler_Withdrawals() {
	mockWithdrawService := &MockWithdrawRepository{
		WithdrawalsFunc: func() ([]interfaces.WithdrawInfo, error) {
			return []interfaces.WithdrawInfo{
				{OrderNumber: "123456", Sum: 100},
			}, nil
		},
	}

	userHandler := UserHandler{
		WithdrawService: mockWithdrawService,
	}

	req := httptest.NewRequest("GET", "/api/user/withdrawals", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, 123))
	rr := httptest.NewRecorder()

	http.HandlerFunc(userHandler.Withdrawals).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		return
	}
}

func ExampleUserHandler_GetOrders() {
	mockOrderService := &MockOrderService{
		GetUserOrdersFunc: func(userID int) ([]interfaces.OrderData, error) {
			return []interfaces.OrderData{
				{Number: "123456", Status: "Processed"},
			}, nil
		},
	}

	userHandler := UserHandler{
		OrderService: mockOrderService,
	}

	req := httptest.NewRequest("GET", "/api/user/orders", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, 123))
	rr := httptest.NewRecorder()

	http.HandlerFunc(userHandler.GetOrders).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		return
	}
}
