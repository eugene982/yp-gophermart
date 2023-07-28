package config

import (
	"flag"

	"github.com/caarlos0/env/v8"
)

// Объявление структуры конфигурации
type Configuration struct {
	ServAddr             string `env:"RUN_ADDRESS"`            // адрес сервера
	DatabaseDSN          string `env:"DATABASE_URI"`           // адрес подключения к базе данных
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"` // адрес системы расчёта начислений

	Timeout  int    `env:"SERVER_TIMEOUT"` // таймаут сервера
	LogLevel string `env:"LOG_LEVEL"`      // уровень логирования
}

// Возвращаем копию конфигурации полученную из флагов и окружения
func Config() Configuration {
	var config Configuration

	// устанавливаем переменные для флага по умолчанию
	flag.StringVar(&config.ServAddr, "a", "localhost:8080", "server address")
	flag.IntVar(&config.Timeout, "t", 30, "timeout in seconds")
	flag.StringVar(&config.LogLevel, "l", "info", "log level")
	//flag.StringVar(&config.AccrualSystemAddress, "r", "http://localhost:8090", "accural system address")
	flag.StringVar(&config.AccrualSystemAddress, "r", "", "accural system address")
	//flag.StringVar(&config.DatabaseDSN, "d", "postgres://postgres:postgres@localhost/yp_gophermart", "postgres connection string")
	flag.StringVar(&config.DatabaseDSN, "d", "", "postgres connection string")

	// получаем конфигурацию из флагов и/или окружения
	flag.Parse()
	env.Parse(&config)
	return config
}
