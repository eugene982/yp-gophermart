package application

import (
	"net/http"

	"github.com/eugene982/yp-gophermart/internal/logger"
)

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
