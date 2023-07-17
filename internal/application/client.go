package application

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
	accrueReqestDuration = time.Second * 5
	updateOrders         = []string{"NEW", "REGISTERED", "PROCESSING"}
	updateLimit          = 10
)

var errTooManyRequests = errors.New("too many requests")

func (a *Application) startAccrualReqestAsync() {
	ticker := time.NewTicker(accrueReqestDuration)

	for range ticker.C { // пауза между вызовами

		logger.Info("start reqest accrual", "address", a.accrualSystem)
		ctx := context.Background()
		err := a.accuralRequest(ctx)
		logger.Info("end request accrual", "error", err)

	}
}

// Опрос сервиса начисления бонусов
func (a *Application) accuralRequest(ctx context.Context) error {

	orders, err := a.storage.ReadOrdersWithStatus(ctx, updateOrders, updateLimit)
	if err != nil {
		return err
	}

	updOrders := make([]model.OrderInfo, 0)
	updAccruals := make([]model.LoyaltyInfo, 0)

	for _, o := range orders {
		resp, err := a.accrualRequestOrder(ctx, o.OrderID)
		if err != nil {
			if errors.Is(err, errTooManyRequests) {
				break
			}
			return err
		} else if resp.Order == "" {
			continue
		}

		o.Status = strings.ToUpper(resp.Status)
		updOrders = append(updOrders, o)

		if resp.Accrual >= 0.01 || resp.Accrual <= -0.01 {
			updAccruals = append(updAccruals, model.LoyaltyInfo{
				UserID:     o.UserID,
				OrderID:    o.OrderID,
				IsAccrual:  true,
				Points:     resp.Accrual,
				UploadedAt: time.Now(),
			})
		}
	}
	return a.storage.UpdateOrders(ctx, updOrders, updAccruals)
}

// запрос к внешней системе по отдельному заказу
func (a *Application) accrualRequestOrder(ctx context.Context, orderID int64) (res model.AccrualResponse, err error) {
	select {
	case <-ctx.Done():
		return res, ctx.Err()
	default:
	}

	r, err := a.client.Get(fmt.Sprintf("%s/api/orders/%d", a.accrualSystem, orderID))
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
