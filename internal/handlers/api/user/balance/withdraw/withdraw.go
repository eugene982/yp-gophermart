package withdraw

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/eugene982/yp-gophermart/internal/handlers"
	"github.com/eugene982/yp-gophermart/internal/logger"
	"github.com/eugene982/yp-gophermart/internal/middleware"
	"github.com/eugene982/yp-gophermart/internal/model"
	"github.com/eugene982/yp-gophermart/internal/utils"
)

type WithdrawReadWtiter interface {
	handlers.BalanceReader
	handlers.WithdrawWriter
}

// запрос на списание средств
func NewWithdrawHandler(rw WithdrawReadWtiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			logger.Info("invalid header", "Content-Type", contentType)
			http.Error(w, "invalid content-type", http.StatusBadRequest)
			return
		}

		userID, err := middleware.GetCookieUserID(r)
		if err != nil {
			logger.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var request model.WithdrawRequest
		err = json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			logger.Info("bad reqest", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		order, err := utils.OrderNumberToInt(request.Order)
		if err != nil {
			logger.Info("invalid order number", "err", err)
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		// проверка наличия достаточного остатка
		if balance, err := rw.ReadBalance(r.Context(), userID); err != nil {
			if handlers.IsNoContent(err) {
				logger.Info("payment required", "balance", err)
				http.Error(w, "402 Payment required", http.StatusPaymentRequired)
			} else {
				logger.Error(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		} else if balance.Current < int(request.Sum*100) {
			logger.Info("payment required", "balance", balance)
			http.Error(w, "402 Payment required", http.StatusPaymentRequired)
			return
		}

		if err = rw.WriteWithdraw(r.Context(), userID, order, int(request.Sum*100)); err != nil {
			logger.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
