package application

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/eugene982/yp-gophermart/internal/logger"
	"github.com/eugene982/yp-gophermart/internal/middleware"
	"github.com/eugene982/yp-gophermart/internal/model"
	"github.com/eugene982/yp-gophermart/internal/storage"
)

// получение остатков баллов пользователя
func (a *Application) getBalanceHandler(w http.ResponseWriter, r *http.Request) {

	userID, err := middleware.GetCookieUserID(r)
	if err != nil {
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// получаем сведения о лояльности
	balance, err := a.storage.ReadBalance(r.Context(), userID)
	if err != nil && !errors.Is(err, storage.ErrNoContent) {
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := model.BalanceResponse{
		Current:   float32(balance.Current) / 100.0,
		Withdrawn: float32(balance.Withdrawn) / 100.0,
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(response); err != nil {
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
