package application

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/eugene982/yp-gophermart/internal/services/clients"
	"github.com/eugene982/yp-gophermart/internal/services/database"
	"github.com/eugene982/yp-gophermart/internal/services/database/postgres"

	"github.com/eugene982/yp-gophermart/internal/config"
	"github.com/eugene982/yp-gophermart/internal/model"
	"github.com/eugene982/yp-gophermart/internal/utils"
)

const (
	// подмешиваем в пароль при получении хеша
	passwordSalt = "YandexPracticumSalt"

	// период между опросами внешнего сервиса
	accrueReqestDuration = time.Second * 5

	// Количество заказов для обновления, получаемые из БД
	updateOrderLimit = 10
)

type PasswdHashFunc func(model.LoginReqest) string

type Application struct {
	storage database.Database // база данных, хранилище
	server  *http.Server      // запускаемый сервер при старте приложения
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
		err = clients.Initialize(time.Second*time.Duration(conf.Timeout),
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

	utils.PasswordHash = func(r model.LoginReqest) string {
		h := sha256.New()
		return fmt.Sprintf("%x",
			h.Sum([]byte(r.Password+passwordSalt+r.Login)))
	}

	return &a, nil
}

func (a *Application) Start() error {

	// Стартуем опрос внешней системы в отдельной горутине
	go clients.StartAccrualReqestAsync(a.storage, accrueReqestDuration)

	return a.server.ListenAndServe()
}

// Освобождение ресурсов приложения
func (a *Application) Close() error {
	err := a.storage.Close()
	if e := a.server.Close(); err != nil && e != nil && !errors.Is(err, http.ErrServerClosed) {
		err = e
	}
	return err
}
