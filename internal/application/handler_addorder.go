package application

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/eugene982/yp-gophermart/internal/logger"
	"github.com/eugene982/yp-gophermart/internal/middleware"
	"github.com/eugene982/yp-gophermart/internal/model"
	"github.com/eugene982/yp-gophermart/internal/storage"
)

// Запись данных заказа в хранилище
func (a *Application) addOrderHandler(w http.ResponseWriter, r *http.Request) {

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
	order, err := orderNumberToInt(string(body))
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

	data := model.OrderInfo{
		UserID:     userID,
		OrderID:    order,
		Status:     "NEW",
		UploadedAt: time.Now(),
	}

	if err = a.storage.WriteOrder(r.Context(), data); err != nil {
		if !errors.Is(err, storage.ErrWriteConflict) {
			logger.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		logger.Info("write order conflict", "error", err, "number", order)

		// проверяем кому принадлежит номер
		userOrders, err := a.storage.ReadOrders(r.Context(), userID, order)
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
