package application

import (
	"errors"
	"net/http"
	"time"

	"github.com/eugene982/yp-gophermart/internal/config"
	"github.com/eugene982/yp-gophermart/internal/services/clients"

	"github.com/eugene982/yp-gophermart/internal/services/database"
	_ "github.com/eugene982/yp-gophermart/internal/services/database/postgres" // чтоб init() отработал
)

const (
	// подмешиваем в пароль при получении хеша
	passwordSalt = "YandexPracticumSalt"

	// период между опросами внешнего сервиса
	accrueReqestDuration = time.Second * 5

	// Количество заказов для обновления, получаемые из БД
	updateOrderLimit = 10
)

type Application struct {
	storage database.Database      // база данных, хранилище
	server  *http.Server           // запускаемый сервер при старте приложения
	client  *clients.AccrualClient // клиент опроса внешней системы
}

// Создание экземпляра приложения
func New(conf config.Configuration) (*Application, error) {

	var (
		err error
		a   Application
	)

	// установка соединения БД
	a.storage, err = database.Open(conf.DatabaseDSN)
	if err != nil {
		return nil, err
	}

	// клиент, который опрашивает внешний ресурс
	if conf.AccrualSystemAddress != "" {
		a.client, err = clients.NewAccrualClient(time.Second*time.Duration(conf.Timeout),
			conf.AccrualSystemAddress, updateOrderLimit)
		if err != nil {
			return nil, err
		}
	}

	a.server = &http.Server{
		Addr:         conf.ServAddr,
		WriteTimeout: time.Second * time.Duration(conf.Timeout),
		ReadTimeout:  time.Second * time.Duration(conf.Timeout),
		Handler:      newRouter(a.storage),
	}

	return &a, nil
}

func (a *Application) Start() error {
	// Стартуем опрос внешней системы в отдельной горутине
	if a.client != nil {
		a.client.StartReqestAsync(a.storage, accrueReqestDuration)
	}

	return a.server.ListenAndServe()
}

// Освобождение ресурсов приложения
func (a *Application) Close() error {
	if a.client != nil {
		a.client.Stop()
	}

	err := a.storage.Close()
	if e := a.server.Close(); err != nil && e != nil && !errors.Is(err, http.ErrServerClosed) {
		err = e
	}
	return err
}
