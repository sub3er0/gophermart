package config

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	ServerAddress        string
	DatabaseDsn          string
	AccrualSystemAddress string
}

func InitConfig() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8088", "Адрес HTTP-сервера")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "http://localhost:8080", "Адрес системы расчета")
	flag.StringVar(
		&cfg.DatabaseDsn,
		"b", "postgres://postgres:326717@localhost:5432/gofermart?sslmode=disable",
		"Строка подключения к базе данных")
	flag.Parse()

	if ServerAddress := os.Getenv("RUN_ADDRESS"); ServerAddress != "" {
		cfg.ServerAddress = ServerAddress
	}

	if DatabaseDsn := os.Getenv("DATABASE_URI"); DatabaseDsn != "" {
		cfg.DatabaseDsn = DatabaseDsn
	}

	if AccrualSystemAddress := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); AccrualSystemAddress != "" {
		cfg.AccrualSystemAddress = AccrualSystemAddress
	}

	if cfg.ServerAddress == "" {
		return nil, fmt.Errorf("ServerAddress is required")
	}

	return cfg, nil
}
