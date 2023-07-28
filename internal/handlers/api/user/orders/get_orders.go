package orders

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/eugene982/yp-gophermart/internal/handlers"
	"github.com/eugene982/yp-gophermart/internal/logger"
	"github.com/eugene982/yp-gophermart/internal/middleware"
	"github.com/eugene982/yp-gophermart/internal/model"
)

type OrderAcrualReader interface {
	handlers.OrderReader
	handlers.AccrualReader
}

// чтедине данных заказа пользователя
func NewGetOrdersHandler(reader OrderAcrualReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userID, err := middleware.GetCookieUserID(r)
		if err != nil {
			logger.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		orders, err := reader.ReadOrders(r.Context(), userID)
		if err != nil {
			logger.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(orders) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// получаем сведения о лояльности
		operations, err := reader.ReadAccruals(r.Context(), userID)
		if err != nil {
			logger.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// сгрупируем по номеру
		accruals := make(map[int64]int)
		for _, l := range operations {
			accruals[l.OrderID] += l.Points
		}

		response := make([]model.OrderResponse, len(orders))
		for i, o := range orders {
			response[i] = model.OrderResponse{
				Number:     strconv.FormatInt(o.OrderID, 10),
				Status:     strings.ToUpper(o.Status),
				Accrual:    float32(accruals[o.OrderID]) / 100,
				UploadedAt: o.UploadedAt.Format(time.RFC3339),
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(response); err != nil {
			logger.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
