package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/eugene982/yp-gophermart/internal/logger"
	"github.com/go-chi/jwtauth/v5"
)

const (
	secretKey = "SecretSignKey"
	tokenExp  = time.Hour * 3
)

var (
	tokenAuth *jwtauth.JWTAuth
)

type contextKeyType uint

const (
	contextKeyUserID contextKeyType = iota
)

func init() {
	// аутентификация
	tokenAuth = jwtauth.New("HS256", []byte(secretKey), nil)
}

// Прослойка аутинтификации пользователя с помощью куки
func CookieAuth(next http.Handler) http.Handler {

	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		_, claims, err := jwtauth.FromContext(ctx)
		// Токен не создат, или истекло время
		if errors.Is(err, jwtauth.ErrNoTokenFound) || errors.Is(err, jwtauth.ErrExpired) {
			logger.Info("unauthorized", "error", err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// 	любая другая ошибка получения токена
		if err != nil {
			logger.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// токен существует, проверка идентификатора пользователя
		id, ok := claims["user_id"]
		if !ok {
			logger.Info("user id not found in claims")
			http.Error(w, "401 Unauthorized", http.StatusUnauthorized)
			return
		}

		userID, ok := id.(string)
		if !ok {
			logger.Error(fmt.Errorf("cannot convert to string"), "user_id", id)
			http.Error(w, "500 Internal server error", http.StatusInternalServerError)
			return
		}

		// положим идентификатор пользователя в контекст, что быстро получать
		ru := r.WithContext(context.WithValue(ctx, contextKeyUserID, userID))
		logger.Info("cookie", "user_id", userID)

		next.ServeHTTP(w, ru)
	}

	// запускаем через верификатор
	return jwtauth.Verifier(tokenAuth)(
		http.HandlerFunc(fn))
}

// Добавление идентификатора пользователя в куки
func SetCookieUserID(userID string, w http.ResponseWriter) error {
	_, tokenString, err := tokenAuth.Encode(map[string]interface{}{
		"user_id": userID,
	})
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "jwt",
		Value:   tokenString,
		Expires: time.Now().Add(tokenExp),
	})
	return nil
}

// Возвращает идентификатор пользователя из контекста
func GetCookieUserID(r *http.Request) (string, error) {
	val := r.Context().Value(contextKeyUserID)
	if val == nil {
		return "", fmt.Errorf("user id not found")
	}
	userID, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("user id is not uint type")
	}
	return userID, nil
}
