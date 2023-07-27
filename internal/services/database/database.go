package database

//go:generate go run github.com/vektra/mockery/v2@v2.32.0 --name=Database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/eugene982/yp-gophermart/internal/model"
	"github.com/jmoiron/sqlx"
)

var (
	ErrWriteConflict = errors.New("data exists")
	ErrNoContent     = errors.New("no content")
	ErrDBNotInit     = errors.New("database not initialize")
)

var database Database

func Open(dns string) (Database, error) {

	if dns == "" {
		return nil, fmt.Errorf("database dsn is empty")
	}

	db, err := sqlx.Open("pgx", dns)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	// Настройка пула соединений
	db.SetMaxOpenConns(3)
	db.SetMaxIdleConns(3)
	db.SetConnMaxLifetime(3 * time.Minute)

	if database == nil {
		return nil, ErrDBNotInit
	}

	err = database.Open(db)
	if err != nil {
		return nil, err
	}

	return database, nil
}

// регистрируем подключенный драйвер
func RegDriver(db Database) {
	database = db
}

// Интерфейс для хранилища данных
type Database interface {
	Open(*sqlx.DB) error
	Close() error
	Ping(context.Context) error

	WriteUser(ctx context.Context, data model.UserInfo) error
	ReadUser(ctx context.Context, userID string) (model.UserInfo, error)

	WriteNewOrder(ctx context.Context, userID string, order int64) error
	ReadOrders(ctx context.Context, userID string, orders ...int64) ([]model.OrderInfo, error)

	WriteWithdraw(ctx context.Context, userID string, order int64, sum int) error
	ReadWithdraws(ctx context.Context, userID string) ([]model.OperationsInfo, error)

	ReadAccruals(ctx context.Context, userID string) ([]model.OperationsInfo, error)

	ReadBalance(ctx context.Context, userID string) (model.BalanceInfo, error)
	ReadOrdersWithStatus(ctx context.Context, status []string, limit int) ([]model.OrderInfo, error)
	UpdateOrderAccrual(ctx context.Context, order model.OrderInfo, accrual int) error
}
