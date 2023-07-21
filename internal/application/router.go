package application

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"

	"github.com/eugene982/yp-gophermart/internal/logger"
	"github.com/eugene982/yp-gophermart/internal/middleware"
)

// Возвращает роутер
func newRouter(a *Application) http.Handler {

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
