package clients

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/eugene982/yp-gophermart/internal/logger"
	"github.com/eugene982/yp-gophermart/internal/model"
)

var (
	client           *http.Client
	updateOrderLimit int
	systemAddress    string
)

var (
	updateOrders       = []string{"NEW", "REGISTERED", "PROCESSING"}
	errTooManyRequests = errors.New("too many requests")
)

// Инициализация клиента
func Initialize(timeout time.Duration, address string, limit int) error {
	if address == "" {
		return fmt.Errorf("acrual system address is emty")
	}
	systemAddress = address

	client = &http.Client{
		Timeout: timeout,
	}
	return nil
}

type OrdersReadWriter interface {
	ReadOrdersWithStatus(ctx context.Context, status []string, limit int) ([]model.OrderInfo, error)
	UpdateOrderAccrual(ctx context.Context, order model.OrderInfo, accrual int) error
}

// Запуск опроса клиентом внешнего сервиса
func StartAccrualReqestAsync(rw OrdersReadWriter, duration time.Duration) {
	if client == nil {
		return
	}

	ticker := time.NewTicker(duration)

	for range ticker.C { // пауза между вызовами

		logger.Info("start reqest accrual", "address", systemAddress)
		err := updateOrdersStatuses(rw)
		logger.Info("end request accrual", "error", err)
	}
}

// Опрос сервиса начисления бонусов
func updateOrdersStatuses(rw OrdersReadWriter) error {
	ctx := context.Background()

	orders, err := rw.ReadOrdersWithStatus(ctx, updateOrders, updateOrderLimit)
	if err != nil {
		return err
	}

	for _, o := range orders {
		resp, err := accrualRequestOrder(ctx, o.OrderID)
		if err != nil {
			if errors.Is(err, errTooManyRequests) {
				break
			}
			return err
		} else if resp.Order == "" {
			continue
		}

		o.Status = strings.ToUpper(resp.Status)
		accrual := int(resp.Accrual * 100)
		rw.UpdateOrderAccrual(ctx, o, accrual)

	}
	return nil
}

// запрос к внешней системе по отдельному заказу
func accrualRequestOrder(ctx context.Context, orderID int64) (res model.AccrualResponse, err error) {
	select {
	case <-ctx.Done():
		return res, ctx.Err()
	default:
	}

	r, err := client.Get(fmt.Sprintf("%s/api/orders/%d", systemAddress, orderID))
	if err != nil {
		return
	}
	defer r.Body.Close()

	switch r.StatusCode {
	case http.StatusOK:
		err = json.NewDecoder(r.Body).Decode(&res)
		return

	case http.StatusNoContent:
		return

	case http.StatusTooManyRequests:
		return res, errTooManyRequests
	}

	body, _ := io.ReadAll(r.Body)
	return res, fmt.Errorf("fail request %s %s", r.Status, string(body))
}
