package middleware

import (
	"net/http"
	"time"

	"github.com/eugene982/yp-gophermart/internal/logger"
)

// структура захвата Ответа сервиса для логирования
type logResponseWriter struct {
	http.ResponseWriter
	size       int
	statusCode int
}

// Проверка на тип, на всякий...
var _ http.ResponseWriter = (*logResponseWriter)(nil)

// Write implements http.ResponseWriter
func (l *logResponseWriter) Write(bytes []byte) (size int, err error) {
	size, err = l.ResponseWriter.Write(bytes)
	l.size += size
	return
}

// WriteHeader implements http.ResponseWriter
func (l *logResponseWriter) WriteHeader(statusCode int) {
	l.ResponseWriter.WriteHeader(statusCode)
	l.statusCode = statusCode
}

// Логирование запросов
func Logger(next http.Handler) http.Handler {

	fn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// обернём записывальщик
		logWriter := &logResponseWriter{w, 0, 0}

		logger.Info(
			"incoming request",
			"method", r.Method,
			"path", r.URL.Path,
			"content_type", r.Header.Get("Content-Type"),
			"accept_encoding", r.Header.Get("Accept-Encoding"),
			"content_encoding", r.Header.Get("Content-Encoding"),
		)

		next.ServeHTTP(logWriter, r)

		logger.Info(
			"outgoing response",
			"status_code", logWriter.statusCode,
			"size", logWriter.size,
			"duration", time.Since(start).String(),
			"content_encoding", w.Header().Get("Content-Encoding"),
		)
	}

	return http.HandlerFunc(fn)
}
