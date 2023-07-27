package register

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/eugene982/yp-gophermart/internal/handlers"
	"github.com/eugene982/yp-gophermart/internal/logger"
	"github.com/eugene982/yp-gophermart/internal/middleware"
	"github.com/eugene982/yp-gophermart/internal/model"
)

// регистрация пользователя
func NewRegisterHandler(writer handlers.UserWriter, hasher handlers.PasswordHasher) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			logger.Info("invalid header", "Content-Type", contentType)
			http.Error(w, "invalid content-type", http.StatusBadRequest)
			return
		}

		var request model.LoginReqest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			logger.Info("bad reqest", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if ok, err := request.IsValid(); !ok {
			logger.Info("bad request", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// хешируем пароли чтоб не хранить их в открытом виде
		userInfo := model.UserInfo{
			UserID:       request.Login,
			PasswordHash: hasher.Hash(request),
		}

		err = writer.WriteUser(r.Context(), userInfo)
		if err != nil {
			if handlers.IsWriteConflict(err) {
				logger.Info("user conflict",
					"error", err,
					"login", request.Login)
				http.Error(w, "user conflict", http.StatusConflict)
			} else {
				logger.Error(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// запоминаем пользователя в куках
		err = middleware.SetCookieUserID(request.Login, w)
		if err != nil {
			logger.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
