package application

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/eugene982/yp-gophermart/internal/logger"
	"github.com/eugene982/yp-gophermart/internal/middleware"
	"github.com/eugene982/yp-gophermart/internal/model"
	"github.com/eugene982/yp-gophermart/internal/storage"
)

// // вход пользователя
func (a *Application) loginUserHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	contentType := r.Header.Get("Content-type")
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

	// получаем данные пользователя и проверяем пароль
	userInfo, err := a.storage.ReadUser(r.Context(), request.Login)
	if err != nil {
		if errors.Is(err, storage.ErrNoContent) {
			logger.Info("user not found", "login", request.Login)
			http.Error(w, "user not found", http.StatusUnauthorized)
		} else {
			logger.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if userInfo.PasswordHash != a.passwdHash(request) {
		logger.Info("password does not match",
			"login", request.Login)
		http.Error(w, "password does not match", http.StatusUnauthorized)
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
