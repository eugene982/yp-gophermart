package model

import (
	"errors"
	"strings"
)

// структура регистрации пользователя
type LoginReqest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// валидация данных пользователя
func (r LoginReqest) IsValid() (bool, error) {
	if strings.TrimSpace(r.Login) == "" {
		return false, errors.New("login is empty")
	}
	if strings.TrimSpace(r.Password) == "" {
		return false, errors.New("password is empty")
	}
	return true, nil
}

// структура ответа заказа
type OrderResponse struct {
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accrual    float32 `json:"accrual,omitempty"`
	UploadedAt string  `json:"uploaded_at"`
}

// структура ответа баланса баллов
type BalanceResponse struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

// структура запроса на списание средств
type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float32 `json:"sum"`
}

// структура ответа о списании средств
type WithdrawResponse struct {
	Order       string  `json:"order"`
	Sum         float32 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

// структура ответа внешнего сервиса
type AccrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual,omitempty"`
}
