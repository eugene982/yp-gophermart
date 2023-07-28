package balance

import (
	"encoding/json"
	"net/http"

	"github.com/eugene982/yp-gophermart/internal/handlers"
	"github.com/eugene982/yp-gophermart/internal/logger"
	"github.com/eugene982/yp-gophermart/internal/middleware"
	"github.com/eugene982/yp-gophermart/internal/model"
)

// получение остатков баллов пользователя
func NewBalanceHandler(reader handlers.BalanceReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userID, err := middleware.GetCookieUserID(r)
		if err != nil {
			logger.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// получаем сведения о лояльности
		balance, err := reader.ReadBalance(r.Context(), userID)
		if err != nil && !handlers.IsNoContent(err) {
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
}
