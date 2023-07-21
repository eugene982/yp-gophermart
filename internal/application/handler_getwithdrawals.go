package application

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/eugene982/yp-gophermart/internal/logger"
	"github.com/eugene982/yp-gophermart/internal/middleware"
	"github.com/eugene982/yp-gophermart/internal/model"
)

// чтедине данных заказа пользователя
func (a *Application) getWithdrawalsHandler(w http.ResponseWriter, r *http.Request) {

	userID, err := middleware.GetCookieUserID(r)
	if err != nil {
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// все данные лояльности
	operations, err := a.storage.ReadOperations(r.Context(), userID, false)
	if err != nil {
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(operations) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response := make([]model.WithdrawResponse, len(operations))
	for i, l := range operations {
		response[i] = model.WithdrawResponse{
			Order:       strconv.FormatInt(l.OrderID, 10),
			Sum:         float32(l.Points) / 100.0,
			ProcessedAt: l.UploadedAt.Format(time.RFC3339),
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
