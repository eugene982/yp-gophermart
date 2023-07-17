package model

import (
	"time"
)

// структура записи пользователя
type UserInfo struct {
	UserID       string `db:"user_id"`
	PasswordHash string `db:"passwd_hash"`
}

// структура записи заказа
type OrderInfo struct {
	UserID     string    `db:"user_id"`
	OrderID    int64     `db:"order_id"`
	Status     string    `db:"status"`
	UploadedAt time.Time `db:"uploaded_at"`
}

// структура записи данных дояльности
type LoyaltyInfo struct {
	UserID     string    `db:"user_id"`
	OrderID    int64     `db:"order_id"`
	IsAccrual  bool      `db:"is_accrual"`
	Points     float32   `db:"points"`
	UploadedAt time.Time `db:"uploaded_at"`
}

// структура ответа баланса баллов
type BalanceInfo struct {
	UserID    string  `db:"user_id"`
	Current   float32 `db:"current"`
	Withdrawn float32 `db:"withdrawn"`
}
