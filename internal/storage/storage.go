package storage

import (
	"context"
	"errors"

	"github.com/eugene982/yp-gophermart/internal/model"
)

var (
	ErrWriteConflict = errors.New("data exists")
)

// Интерфейс для хранилища данных
type Storage interface {
	Close() error
	Ping(context.Context) error

	WriteUser(ctx context.Context, data model.UserInfo) error
	ReadUsers(ctx context.Context, userIDs ...string) ([]model.UserInfo, error)

	WriteOrder(ctx context.Context, data model.OrderInfo) error
	ReadOrders(ctx context.Context, userID string, nums ...int64) ([]model.OrderInfo, error)

	WriteLoyalty(ctx context.Context, data []model.LoyaltyInfo) error
	ReadLoyalty(ctx context.Context, userID string, accrual bool) ([]model.LoyaltyInfo, error)

	ReadBalances(ctx context.Context, userIDs ...string) ([]model.BalanceInfo, error)
	ReadOrdersWithStatus(ctx context.Context, status []string, limit int) ([]model.OrderInfo, error)
	UpdateOrders(ctx context.Context, orders []model.OrderInfo, accrues []model.LoyaltyInfo) error
}
