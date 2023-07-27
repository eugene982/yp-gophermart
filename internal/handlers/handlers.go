package handlers

import (
	"context"
	"errors"

	"github.com/eugene982/yp-gophermart/internal/model"
	"github.com/eugene982/yp-gophermart/internal/services/database"
)

type Pinger interface {
	Ping(context.Context) error
}

type UserWriter interface {
	WriteUser(ctx context.Context, data model.UserInfo) error
}

type UserReader interface {
	ReadUser(ctx context.Context, userID string) (model.UserInfo, error)
}

type OrderWriter interface {
	WriteNewOrder(ctx context.Context, userID string, order int64) error
}

type OrderReader interface {
	ReadOrders(ctx context.Context, userID string, orders ...int64) ([]model.OrderInfo, error)
}

type AccrualReader interface {
	ReadAccruals(ctx context.Context, userID string) ([]model.OperationsInfo, error)
}

type BalanceReader interface {
	ReadBalance(ctx context.Context, userID string) (model.BalanceInfo, error)
}

type WithdrawWriter interface {
	WriteWithdraw(ctx context.Context, userID string, order int64, sum int) error
}

type WithdrawReader interface {
	ReadWithdraws(ctx context.Context, userID string) ([]model.OperationsInfo, error)
}

type PasswordHasher interface {
	Hash(model.LoginReqest) string
}

func IsWriteConflict(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, database.ErrWriteConflict)
}

func IsNoContent(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, database.ErrNoContent)
}
