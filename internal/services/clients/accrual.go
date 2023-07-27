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
	updateOrders       = []string{"NEW", "REGISTERED", "PROCESSING"}
	errTooManyRequests = errors.New("too many requests")
)

type AccrualClient struct {
	client           *http.Client
	updateOrderLimit int
	systemAddress    string
	stopChan         chan struct{}
}

// Инициализация клиента
func NewAccrualClient(timeout time.Duration, address string, limit int) (*AccrualClient, error) {
	if address == "" {
		return nil, fmt.Errorf("acrual system address is emty")
	}
	ac := AccrualClient{
		systemAddress:    address,
		updateOrderLimit: limit,
		client: &http.Client{
			Timeout: timeout,
		},
	}

	return &ac, nil
}

type OrdersReadWriter interface {
	ReadOrdersWithStatus(ctx context.Context, status []string, limit int) ([]model.OrderInfo, error)
	UpdateOrderAccrual(ctx context.Context, order model.OrderInfo, accrual int) error
}

// Запуск опроса клиентом внешнего сервиса
func (ac *AccrualClient) StartReqestAsync(rw OrdersReadWriter, duration time.Duration) {
	ac.stopChan = make(chan struct{})
	ticker := time.NewTicker(duration)

	go func(tick <-chan time.Time) {
		for { // пауза между вызовами

			select {
			case <-ac.stopChan:
				logger.Info("acrual client stop")
			case <-tick:
				logger.Info("start reqest accrual", "address", ac.systemAddress)
				count, err := ac.updateOrdersStatuses(rw)
				if err != nil {
					logger.Warn("end request accrual", "error", err, "update", count)
				} else {
					logger.Info("end request accrual", "update", count)
				}
			}
		}
	}(ticker.C)
}

func (ac *AccrualClient) Stop() {
	if ac.stopChan != nil {
		ac.stopChan <- struct{}{}
	}
}

// Опрос сервиса начисления бонусов
func (ac *AccrualClient) updateOrdersStatuses(rw OrdersReadWriter) (int, error) {
	ctx := context.Background()

	orders, err := rw.ReadOrdersWithStatus(ctx, updateOrders, ac.updateOrderLimit)
	if err != nil {
		return 0, err
	}

	var count int
	for _, o := range orders {
		resp, err := ac.accrualRequestOrder(ctx, o.OrderID)
		if err != nil {
			if errors.Is(err, errTooManyRequests) {
				break
			}
			return count, err
		} else if resp.Order == "" {
			continue
		}

		o.Status = strings.ToUpper(resp.Status)
		accrual := int(resp.Accrual * 100)
		err = rw.UpdateOrderAccrual(ctx, o, accrual)
		if err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

// запрос к внешней системе по отдельному заказу
func (ac *AccrualClient) accrualRequestOrder(ctx context.Context, orderID int64) (res model.AccrualResponse, err error) {
	select {
	case <-ctx.Done():
		return res, ctx.Err()
	default:
	}

	r, err := ac.client.Get(fmt.Sprintf("%s/api/orders/%d", ac.systemAddress, orderID))
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
