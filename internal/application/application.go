package application

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/eugene982/yp-gophermart/internal/config"
	"github.com/eugene982/yp-gophermart/internal/model"
	"github.com/eugene982/yp-gophermart/internal/storage"
	"github.com/eugene982/yp-gophermart/internal/storage/pgstorage"
	"github.com/jmoiron/sqlx"
)

const (
	// подмешиваем в пароль при получении хеша
	passwordSalt         = "YandexPracticumSalt"
	accrueReqestDuration = time.Second * 5
	updateOrderLimit     = 10
)

type PasswdHashFunc func(model.LoginReqest) string

type Application struct {
	accrualSystem string          // адрес системы расчёта начислений
	storage       storage.Storage // хранилище данных
	client        *http.Client
	server        *http.Server   // клиент опрашивающий внешнюю систему
	passwdHash    PasswdHashFunc // функция хеширования пароля
}

func New(conf config.Configuration) (*Application, error) {

	var a Application

	a.accrualSystem = conf.AccrualSystemAddress

	// установка соединения БД
	if conf.DatabaseDSN == "" {
		return nil, fmt.Errorf("database dsn is empty")
	}

	var db *sqlx.DB
	db, err := sqlx.Open("pgx", conf.DatabaseDSN)
	if err != nil {
		return nil, err
	}
	a.storage, err = pgstorage.New(db)
	if err != nil {
		return nil, err
	}

	// клиент, который опрашивает внешний ресурс
	a.client = &http.Client{
		Timeout: time.Second * time.Duration(conf.Timeout),
	}

	a.server = &http.Server{
		Addr:         conf.ServAddr,
		WriteTimeout: time.Second * time.Duration(conf.Timeout),
		ReadTimeout:  time.Second * time.Duration(conf.Timeout),
		Handler:      newRouter(&a),
	}

	a.passwdHash = func(r model.LoginReqest) string {
		h := sha256.New()
		return fmt.Sprintf("%x",
			h.Sum([]byte(r.Password+passwordSalt+r.Login)))
	}

	return &a, nil
}

func (a *Application) Start() error {
	// Стартуем опрос внешней системы в отдельной горутине
	if a.accrualSystem != "" {
		go a.startAccrualReqestAsync()
	}

	//
	srvErr := make(chan error)
	go func() {
		srvErr <- a.server.ListenAndServe()
	}()

	select {
	case err := <-srvErr:
		return err
	default:
		return nil
	}
}

// Освобождение ресурсов приложения
func (a *Application) Close() error {
	err := a.storage.Close()
	if e := a.server.Close(); err != nil && e != nil && !errors.Is(err, http.ErrServerClosed) {
		err = e
	}
	return err
}
