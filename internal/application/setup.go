package application

import (
	"net/http"

	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"

	"github.com/eugene982/yp-gophermart/internal/logger"
	"github.com/eugene982/yp-gophermart/internal/middleware"
	"github.com/eugene982/yp-gophermart/internal/services/database"

	"github.com/eugene982/yp-gophermart/internal/handlers/api/user/balance"
	"github.com/eugene982/yp-gophermart/internal/handlers/api/user/balance/withdraw"
	"github.com/eugene982/yp-gophermart/internal/handlers/api/user/login"
	"github.com/eugene982/yp-gophermart/internal/handlers/api/user/orders"
	"github.com/eugene982/yp-gophermart/internal/handlers/api/user/register"
	"github.com/eugene982/yp-gophermart/internal/handlers/api/user/withdrawals"
	"github.com/eugene982/yp-gophermart/internal/handlers/ping"
)

// Возвращает роутер
func newRouter(db database.Database) http.Handler {

	r := chi.NewRouter()

	r.Use(middleware.Logger)                            // прослойка логирования
	r.Use(chimiddleware.Compress(3, "gzip", "deflate")) // прослойка сжатия

	// методы доступные без авторизации
	r.Group(func(r chi.Router) {
		r.Get("/ping", ping.NewPingHandler(db))
		r.Post("/api/user/register", register.NewRegisterHandler(db))
		r.Post("/api/user/login", login.NewLoginHandler(db))
	})

	// методы доступные с авторизацией
	r.Group(func(r chi.Router) {
		r.Use(middleware.CookieAuth)

		r.Post("/api/user/orders", orders.NewAddOrderHandler(db))
		r.Get("/api/user/orders", orders.NewGetOrdersHandler(db))
		r.Get("/api/user/balance", balance.NewBalanceHandler(db))
		r.Post("/api/user/balance/withdraw", withdraw.NewWithdrawHandler(db))
		r.Get("/api/user/withdrawals", withdrawals.NewWithdrawalsHandler(db))
	})

	// во всех остальных случаях 404
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("not allowed",
			"method", r.Method)
		http.NotFound(w, r)
	})

	return r
}
