package main

import (
	"github.com/go-chi/chi/v5"
	"gophermart/internal/config"
	"gophermart/internal/handlers"
	"gophermart/internal/middleware"
	"gophermart/internal/repository"
	"gophermart/internal/service"
	"gophermart/storage"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net/http"
)

func main() {
	pgsStorage := &storage.PgStorage{}
	cfg, err := config.InitConfig()

	if err != nil {
		log.Fatalf("Error while initializing config: %v", err)
	}

	if cfg != nil {
		dsn := cfg.DatabaseDsn
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

		if err != nil {
			log.Fatalf("failed to connect database: %v", err)
		}

		err = db.AutoMigrate()

		if err != nil {
			log.Fatalf("failed to migrate database: %v", err)
		}

		err = db.AutoMigrate(&storage.Order{})

		if err != nil {
			log.Fatalf("failed to migrate database: %v", err)
		}

		err = db.AutoMigrate(&storage.User{})

		if err != nil {
			log.Fatalf("failed to migrate database: %v", err)
		}

		err = db.AutoMigrate(&storage.UserBalance{})

		if err != nil {
			log.Fatalf("failed to migrate database: %v", err)
		}

		err = db.AutoMigrate(&storage.Withdrawal{})

		if err != nil {
			log.Fatalf("failed to migrate database: %v", err)
		}
	} else {
		log.Fatalf("Error while initializing db connection: %v", err)
	}

	err = pgsStorage.Init(cfg.DatabaseDsn)

	if err != nil {
		log.Fatalf("Error while initializing db connection: %v", err)
	}

	defer pgsStorage.Close()

	userRepository := repository.UserRepository{
		DBStorage: pgsStorage,
	}
	userService := service.UserService{
		UserRepository: &userRepository,
	}
	orderRepository := repository.OrderRepository{
		DBStorage: pgsStorage,
	}
	orderService := service.OrderService{
		OrderRepository: &orderRepository,
	}
	withdrawRepository := repository.WithdrawRepository{
		DBStorage: pgsStorage,
	}
	withdrawService := service.WithdrawService{
		WithdrawRepository: &withdrawRepository,
	}
	userBalanceRepository := repository.UserBalanceRepository{
		DBStorage: pgsStorage,
	}
	userBalanceService := service.UserBalanceService{
		UserBalanceRepository: &userBalanceRepository,
	}
	userHandler := handlers.UserHandler{
		UserService:          &userService,
		OrderService:         &orderService,
		WithdrawService:      &withdrawService,
		UserBalanceService:   &userBalanceService,
		AccrualSystemAddress: cfg.AccrualSystemAddress,
		DbConnectionString:   cfg.DatabaseDsn,
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestDecompressor)
	r.Post("/api/user/register", userHandler.Register)
	r.Post("/api/user/login", userHandler.Login)

	r.With(middleware.TokenAuthMiddleware).Route("/", func(r chi.Router) {
		r.Post("/api/user/orders", userHandler.SaveOrder)
		r.Get("/api/user/orders", userHandler.GetOrders)
		r.Get("/api/user/balance", userHandler.GetBalance)
		r.Post("/api/user/balance/withdraw", userHandler.Withdraw)
		r.Get("/api/user/withdrawals", userHandler.Withdrawals)
	})

	err = http.ListenAndServe(cfg.ServerAddress, r)
	if err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}
