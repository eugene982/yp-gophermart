package application

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"

	"github.com/eugene982/yp-gophermart/internal/logger"
	"github.com/eugene982/yp-gophermart/internal/middleware"
	"github.com/eugene982/yp-gophermart/internal/model"
	"github.com/eugene982/yp-gophermart/internal/storage"
)

// Возвращает роутер
func (a *Application) NewRouter() http.Handler {

	r := chi.NewRouter()

	r.Use(middleware.Logger)                            // прослойка логирования
	r.Use(chimiddleware.Compress(3, "gzip", "deflate")) // прослойка сжатия

	// методы доступные без авторизации
	r.Group(func(r chi.Router) {
		r.Get("/ping", a.pingHandler)
		r.Post("/api/user/register", a.registerUserHadler)
		r.Post("/api/user/login", a.loginUserHandler)
	})

	// методы доступные с авторизацией
	r.Group(func(r chi.Router) {
		r.Use(middleware.CookieAuth)

		r.Post("/api/user/orders", a.addOrderHandler)
		r.Get("/api/user/orders", a.getOrdersHandler)
		r.Get("/api/user/balance", a.getBalanceHandler)
		r.Post("/api/user/balance/withdraw", a.withdrawHandler)
		r.Get("/api/user/withdrawals", a.getWithdrawalsHandler)
	})

	// во всех остальных случаях 404
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("not allowed",
			"method", r.Method)
		http.NotFound(w, r)
	})

	return r
}

// пинг к сервису (бд)
func (a *Application) pingHandler(w http.ResponseWriter, r *http.Request) {

	if err := a.storage.Ping(r.Context()); err != nil {
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

// регистрация пользователя
func (a *Application) registerUserHadler(w http.ResponseWriter, r *http.Request) {
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
		PasswordHash: a.passwdHashFn(request),
	}

	err = a.storage.WriteUser(r.Context(), userInfo)
	if err != nil {
		if errors.Is(err, storage.ErrWriteConflict) {
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

// вход пользователя
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
	usersInfo, err := a.storage.ReadUsers(r.Context(), request.Login)
	if err != nil {
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if len(usersInfo) == 0 {
		logger.Info("user not found", "login", request.Login)
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}

	if usersInfo[0].PasswordHash != a.passwdHashFn(request) {
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

// чтедине данных заказа пользователя
func (a *Application) getOrdersHandler(w http.ResponseWriter, r *http.Request) {

	userID, err := middleware.GetCookieUserID(r)
	if err != nil {
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	orders, err := a.storage.ReadOrders(r.Context(), userID)
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
	loyalty, err := a.storage.ReadLoyalty(r.Context(), userID, true)
	if err != nil {
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// сгрупируем по номеру
	accruals := make(map[int64]float32)
	for _, l := range loyalty {
		accruals[l.OrderID] += l.Points
	}

	response := make([]model.OrderResponse, len(orders))
	for i, o := range orders {
		response[i] = model.OrderResponse{
			Number:     strconv.FormatInt(o.OrderID, 10),
			Status:     strings.ToUpper(o.Status),
			Accrual:    accruals[o.OrderID],
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

// получение остатков баллов пользователя
func (a *Application) getBalanceHandler(w http.ResponseWriter, r *http.Request) {

	userID, err := middleware.GetCookieUserID(r)
	if err != nil {
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response model.BalanceResponse

	// получаем сведения о лояльности
	balances, err := a.storage.ReadBalances(r.Context(), userID)
	if err != nil {
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, b := range balances {
		response.Current += b.Current
		response.Withdrawn += b.Withdrawn
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(response); err != nil {
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// запрос на списание средств
func (a *Application) withdrawHandler(w http.ResponseWriter, r *http.Request) {
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

	order, err := orderNumberToInt(request.Order)
	if err != nil {
		logger.Info("invalid order number", "err", err)
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// проверка наличия достаточного остатка
	if balances, err := a.storage.ReadBalances(r.Context(), userID); err != nil {
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	} else if len(balances) == 0 || balances[0].Current < request.Sum {
		logger.Info("payment required", "balances", balances)
		http.Error(w, "402 Payment required", http.StatusPaymentRequired)
		return
	}

	rec := model.LoyaltyInfo{
		UserID:     userID,
		OrderID:    order,
		IsAccrual:  false,
		Points:     request.Sum,
		UploadedAt: time.Now(),
	}
	if err = a.storage.WriteLoyalty(r.Context(), []model.LoyaltyInfo{rec}); err != nil {
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// чтедине данных заказа пользователя
func (a *Application) getWithdrawalsHandler(w http.ResponseWriter, r *http.Request) {

	userID, err := middleware.GetCookieUserID(r)
	if err != nil {
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// все данные лояльности
	loyalty, err := a.storage.ReadLoyalty(r.Context(), userID, false)
	if err != nil {
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(loyalty) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response := make([]model.WithdrawResponse, len(loyalty))
	for i, l := range loyalty {
		response[i] = model.WithdrawResponse{
			Order:       strconv.FormatInt(l.OrderID, 10),
			Sum:         l.Points,
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

// Проверка корректности номера заказа
func orderNumberToInt(order string) (int64, error) {

	number, err := strconv.ParseInt(order, 10, 64)
	if err != nil {
		return 0, err
	}

	// Valid check number is valid or not based
	// on Luhn algorithm
	var luhn int64
	num := number / 10
	for i := 0; num > 0; i++ {
		cur := num % 10

		if i%2 == 0 { // even
			cur = cur * 2
			if cur > 9 {
				cur = cur%10 + cur/10
			}
		}
		luhn += cur
		num = num / 10
	}

	if (number%10+luhn%10)%10 != 0 {
		return 0, fmt.Errorf("invalid check number %s", order)
	}
	return number, nil
}
