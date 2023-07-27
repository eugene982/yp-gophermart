package application

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/eugene982/yp-gophermart/internal/services/clients"
	"github.com/eugene982/yp-gophermart/internal/services/database"
	"github.com/eugene982/yp-gophermart/internal/services/database/postgres"

	"github.com/eugene982/yp-gophermart/internal/config"
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

	var a Application

	// установка соединения БД
	if conf.DatabaseDSN == "" {
		return nil, fmt.Errorf("database dsn is empty")
	}

	var db *sqlx.DB
	db, err := sqlx.Open("pgx", conf.DatabaseDSN)
	if err != nil {
		return nil, err
	}

	// Переделать !!!
	a.storage, err = postgres.Initialize(db)
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
