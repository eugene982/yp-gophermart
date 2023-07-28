package ping

import (
	"net/http"

	"github.com/eugene982/yp-gophermart/internal/handlers"
	"github.com/eugene982/yp-gophermart/internal/logger"
)

// пинг к сервису (бд)
func NewPingHandler(pinger handlers.Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if err := pinger.Ping(r.Context()); err != nil {
			logger.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	}
}
