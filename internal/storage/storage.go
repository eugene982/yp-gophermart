package storage

import (
	"context"
	"errors"

	"github.com/eugene982/yp-gophermart/internal/model"
)

var (
	ErrWriteConflict = errors.New("data exists")
	ErrNoContent     = errors.New("no content")
)

// Интерфейс для хранилища данных
type Storage interface {
	Close() error
	Ping(context.Context) error

	WriteUser(ctx context.Context, data model.UserInfo) error
	ReadUser(ctx context.Context, userID string) (model.UserInfo, error)

	WriteOrder(ctx context.Context, data model.OrderInfo) error
	ReadOrders(ctx context.Context, userID string, nums ...int64) ([]model.OrderInfo, error)

	WriteOperations(ctx context.Context, data []model.OperationsInfo) error
	ReadOperations(ctx context.Context, userID string, isAccrual bool) ([]model.OperationsInfo, error)

	ReadBalance(ctx context.Context, userID string) (model.BalanceInfo, error)
	ReadOrdersWithStatus(ctx context.Context, status []string, limit int) ([]model.OrderInfo, error)
	UpdateOrderAccrual(ctx context.Context, order model.OrderInfo, accrual int) error
}
