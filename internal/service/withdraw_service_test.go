package service

import (
	"database/sql"
	"errors"
	"github.com/jackc/pgx/v4"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gophermart/internal/interfaces"
	"gophermart/internal/repository"
	"testing"
	"time"
)

// MockWithdrawRepository является мок-реализацией WithdrawRepositoryInterface.
type MockWithdrawRepository struct {
	mock.Mock
}

// GetDBStorage возвращает интерфейс доступа к хранилищу данных.
func (m *MockWithdrawRepository) GetDBStorage() interfaces.DBStorageInterface {
	args := m.Called()
	return args.Get(0).(interfaces.DBStorageInterface)
}

// Withdrawals возвращает список выводов для указанного пользователя.
func (m *MockWithdrawRepository) Withdrawals(userID int) ([]interfaces.WithdrawInfo, error) {
	args := m.Called(userID)
	return args.Get(0).([]interfaces.WithdrawInfo), args.Error(1)
}

// GetCurrentBalance возвращает текущий баланс пользователя.
func (m *MockWithdrawRepository) GetCurrentBalance(userID int) (decimal.Decimal, error) {
	args := m.Called(userID)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

// UpdateUserBalance обновляет баланс пользователя.
func (m *MockWithdrawRepository) UpdateUserBalance(userID int, sum decimal.Decimal) error {
	args := m.Called(userID, sum)
	return args.Error(0)
}

// SaveWithdrawal сохраняет информацию о выводе для указанного пользователя.
func (m *MockWithdrawRepository) SaveWithdrawal(userID int, orderNumber string, sum decimal.Decimal) error {
	args := m.Called(userID, orderNumber, sum)
	return args.Error(0)
}

// MockDBStorage для имитации интерфейса DBStorageInterface
type MockDBStorageWithdraw struct {
	mock.Mock
}

// Init инициализирует соединение с базой данных.
func (m *MockDBStorageWithdraw) Init(connectionString string) error {
	args := m.Called(connectionString)
	return args.Error(0)
}

// Exec выполняет произвольный SQL-запрос с указанными аргументами.
func (m *MockDBStorageWithdraw) Exec(query string, args ...interface{}) (sql.Result, error) {
	// Завершите реализацию в зависимости от ваших требований
	ret := m.Called(query, args)
	return ret.Get(0).(sql.Result), ret.Error(1)
}

// Select выполняет запрос и возвращает строки данных.
func (m *MockDBStorageWithdraw) Select(query string, args ...interface{}) (pgx.Rows, error) {
	// Завершите реализацию в зависимости от ваших требований
	ret := m.Called(query, args)
	return ret.Get(0).(pgx.Rows), ret.Error(1)
}

// Close закрывает соединение с базой данных.
func (m *MockDBStorageWithdraw) Close() {
	m.Called()
}

// BeginTransaction начинает новую транзакцию.
func (m *MockDBStorageWithdraw) BeginTransaction() error {
	args := m.Called()
	return args.Error(0)
}

// Rollback откатывает текущую транзакцию.
func (m *MockDBStorageWithdraw) Rollback() error {
	args := m.Called()
	return args.Error(0)
}

// Commit фиксирует текущую транзакцию.
func (m *MockDBStorageWithdraw) Commit() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDBStorageWithdraw) QueryRow(query string, args ...interface{}) pgx.Row {
	// Завершите реализацию в зависимости от ваших требований
	ret := m.Called(query, args)
	return ret.Get(0).(pgx.Row)
}

func TestWithdraw_Success(t *testing.T) {
	mockRepo := new(MockWithdrawRepository)
	service := WithdrawService{WithdrawRepository: mockRepo}

	userID := 1
	orderNumber := "order-123"
	sum := decimal.NewFromFloat(50.00)

	mockDBStorage := new(MockDBStorageWithdraw)
	mockDBStorage.On("BeginTransaction").Return(nil)
	mockDBStorage.On("Commit").Return(nil)

	mockRepo.On("GetDBStorage").Return(mockDBStorage)
	mockRepo.On("GetCurrentBalance", userID).Return(decimal.NewFromFloat(100.00), nil)
	mockRepo.On("UpdateUserBalance", userID, sum).Return(nil)
	mockRepo.On("SaveWithdrawal", userID, orderNumber, sum).Return(nil)

	code, err := service.Withdraw(userID, orderNumber, sum)

	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	mockRepo.AssertExpectations(t)
	mockDBStorage.AssertExpectations(t)
}

func TestWithdraw_InsufficientFunds(t *testing.T) {
	mockRepo := new(MockWithdrawRepository)
	service := WithdrawService{WithdrawRepository: mockRepo}

	userID := 1
	orderNumber := "order-123"
	sum := decimal.NewFromFloat(100.00)

	mockDBStorage := new(MockDBStorageWithdraw)
	mockDBStorage.On("BeginTransaction").Return(nil)
	mockDBStorage.On("Rollback").Return(nil)
	mockRepo.On("GetDBStorage").Return(mockDBStorage)
	mockRepo.On("GetCurrentBalance", userID).Return(decimal.NewFromFloat(50.00), nil)

	code, err := service.Withdraw(userID, orderNumber, sum)

	assert.NoError(t, err)
	assert.Equal(t, repository.NotEnoughFound, code)

	mockRepo.AssertExpectations(t)
	mockDBStorage.AssertExpectations(t)
}

func TestWithdraw_ErrorGettingBalance(t *testing.T) {
	mockRepo := new(MockWithdrawRepository)
	service := WithdrawService{WithdrawRepository: mockRepo}

	userID := 1
	orderNumber := "order-123"
	sum := decimal.NewFromFloat(50.00)

	mockDBStorage := new(MockDBStorageWithdraw)
	mockDBStorage.On("BeginTransaction").Return(nil)
	mockDBStorage.On("Rollback").Return(nil)
	mockRepo.On("GetDBStorage").Return(mockDBStorage)
	mockRepo.On("GetCurrentBalance", userID).Return(decimal.Decimal{}, errors.New("ошибка получения баланса"))

	code, err := service.Withdraw(userID, orderNumber, sum)

	assert.EqualError(t, err, "ошибка получения баланса")
	assert.Equal(t, repository.WithdrawTransactionError, code)

	mockRepo.AssertExpectations(t)
	mockDBStorage.AssertExpectations(t)
}

func TestWithdrawals(t *testing.T) {
	mockRepo := new(MockWithdrawRepository)
	service := WithdrawService{WithdrawRepository: mockRepo}

	userID := 1
	withdrawals := []interfaces.WithdrawInfo{
		{OrderNumber: "order-123", Sum: 50.00, ProcessedAt: time.Now()},
	}

	mockRepo.On("Withdrawals", userID).Return(withdrawals, nil)

	result, err := service.Withdrawals(userID)

	assert.NoError(t, err)
	assert.Equal(t, withdrawals, result)

	mockRepo.AssertExpectations(t)
}

// Тест для ошибки при обновлении баланса
func TestWithdraw_ErrorUpdatingBalance(t *testing.T) {
	mockRepo := new(MockWithdrawRepository)
	mockDB := new(MockDBStorageWithdraw)
	service := WithdrawService{WithdrawRepository: mockRepo}

	userID := 1
	orderNumber := "order-123"
	sum := decimal.NewFromFloat(50.00)

	// Настройка поведения мока
	mockRepo.On("GetDBStorage").Return(mockDB)
	mockDB.On("BeginTransaction").Return(nil)
	mockRepo.On("GetCurrentBalance", userID).Return(decimal.NewFromFloat(100.00), nil)
	mockRepo.On("UpdateUserBalance", userID, sum).Return(errors.New("ошибка обновления баланса"))
	mockDB.On("Rollback").Return(nil)

	code, err := service.Withdraw(userID, orderNumber, sum)

	assert.EqualError(t, err, "ошибка обновления баланса")
	assert.Equal(t, repository.WithdrawTransactionError, code)

	mockRepo.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

// Тест для ошибки при сохранении вывода
func TestWithdraw_ErrorSavingWithdrawal(t *testing.T) {
	mockRepo := new(MockWithdrawRepository)
	mockDB := new(MockDBStorageWithdraw)
	service := WithdrawService{WithdrawRepository: mockRepo}

	userID := 1
	orderNumber := "order-123"
	sum := decimal.NewFromFloat(50.00)

	// Настройка поведения мока
	mockRepo.On("GetDBStorage").Return(mockDB)
	mockDB.On("BeginTransaction").Return(nil)
	mockRepo.On("GetCurrentBalance", userID).Return(decimal.NewFromFloat(100.00), nil)
	mockRepo.On("UpdateUserBalance", userID, sum).Return(nil)
	mockRepo.On("SaveWithdrawal", userID, orderNumber, sum).Return(errors.New("ошибка сохранения вывода"))
	mockDB.On("Rollback").Return(nil) // Имитация успешного отката

	code, err := service.Withdraw(userID, orderNumber, sum)

	assert.EqualError(t, err, "ошибка сохранения вывода")
	assert.Equal(t, repository.WithdrawTransactionError, code)

	mockRepo.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

// Тест для ошибки при коммите транзакции
func TestWithdraw_ErrorCommittingTransaction(t *testing.T) {
	mockRepo := new(MockWithdrawRepository)
	mockDB := new(MockDBStorageWithdraw)
	service := WithdrawService{WithdrawRepository: mockRepo}

	userID := 1
	orderNumber := "order-123"
	sum := decimal.NewFromFloat(50.00)

	mockRepo.On("GetDBStorage").Return(mockDB)
	mockDB.On("BeginTransaction").Return(nil)
	mockRepo.On("GetCurrentBalance", userID).Return(decimal.NewFromFloat(100.00), nil)
	mockRepo.On("UpdateUserBalance", userID, sum).Return(nil)
	mockRepo.On("SaveWithdrawal", userID, orderNumber, sum).Return(nil)
	mockDB.On("Commit").Return(errors.New("ошибка коммита"))

	code, err := service.Withdraw(userID, orderNumber, sum)

	assert.EqualError(t, err, "ошибка коммита")
	assert.Equal(t, repository.WithdrawTransactionError, code)

	mockRepo.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}
