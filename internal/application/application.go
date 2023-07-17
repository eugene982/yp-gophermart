package application

import (
	"crypto/sha256"
	"fmt"
	"net/http"

	"github.com/eugene982/yp-gophermart/internal/model"
	"github.com/eugene982/yp-gophermart/internal/storage"
)

const (
	// подмешиваем в пароль при получении хеша
	passwordSalt = "YandexPracticumSalt"
)

type Application struct {
	accrualSystem string                         // адрес системы расчёта начислений
	storage       storage.Storage                // хранилище данных
	client        *http.Client                   // клиент опрашивающий внешнюю систему
	passwdHashFn  func(model.LoginReqest) string // функция хеширования пароля
}

// Конструктор приложения
func New(accrualSystem string, store storage.Storage, client *http.Client) (*Application, error) {

	// используем для хеширования пароля алгоритм sha256
	hashFn := func(r model.LoginReqest) string {
		h := sha256.New()
		return fmt.Sprintf("%x",
			h.Sum([]byte(r.Password+passwordSalt+r.Login)))
	}

	app := &Application{
		accrualSystem,
		store,
		client,
		hashFn,
	}

	// Стартуем опрос внешней системы в отдельной горутине
	if accrualSystem != "" {
		go app.startAccrualReqestAsync()
	}

	return app, nil
}

// Освобождение ресурсов приложения
func (a *Application) Close() error {
	return a.storage.Close()
}
