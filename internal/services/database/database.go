package database

//go:generate go run github.com/vektra/mockery/v2@v2.32.0 --name=Database

import (
	"context"
	"errors"

	"github.com/eugene982/yp-gophermart/internal/model"
)

var (
	ErrWriteConflict = errors.New("data exists")
	ErrNoContent     = errors.New("no content")
	ErrDBNotInit     = errors.New("database not initialize")
)

// Интерфейс для хранилища данных
type Database interface {
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
