package orders

import (
	"io"
	"net/http"
	"strings"

	"github.com/eugene982/yp-gophermart/internal/handlers"
	"github.com/eugene982/yp-gophermart/internal/logger"
	"github.com/eugene982/yp-gophermart/internal/middleware"
	"github.com/eugene982/yp-gophermart/internal/utils"
)

type OrderReadWriter interface {
	handlers.OrderReader
	handlers.OrderWriter
}

// Запись данных заказа в хранилище
func NewAddOrderHandler(orw OrderReadWriter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		defer r.Body.Close()

		contentType := r.Header.Get("Content-type")
		if !strings.Contains(contentType, "text/plain") {
			logger.Info("invalid header", "Content-Type", contentType)
			http.Error(w, "invalid content-type", http.StatusBadRequest)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Info("invalid body", "err", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// проверка корректности номера заказа
		order, err := utils.OrderNumberToInt(string(body))
		if err != nil {
			logger.Info("invalid order number", "err", err)
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		userID, err := middleware.GetCookieUserID(r)
		if err != nil {
			logger.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err = orw.WriteNewOrder(r.Context(), userID, order); err != nil {
			if !handlers.IsWriteConflict(err) {
				logger.Error(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			logger.Info("write order conflict", "error", err, "number", order)

			// проверяем кому принадлежит номер
			userOrders, err := orw.ReadOrders(r.Context(), userID, order)
			if err != nil {
				logger.Error(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			} else if len(userOrders) == 0 {
				http.Error(w, "409 Conflict", http.StatusConflict) // существует для другого пользователя
			} else {
				w.WriteHeader(http.StatusOK) // существует для этого пользователя
			}
			return
		}
		w.WriteHeader(http.StatusAccepted) // принят в обработку
	}
}
