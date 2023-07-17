package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/eugene982/yp-gophermart/internal/application"
	"github.com/eugene982/yp-gophermart/internal/config"
	"github.com/eugene982/yp-gophermart/internal/logger"
	"github.com/eugene982/yp-gophermart/internal/storage"
	"github.com/eugene982/yp-gophermart/internal/storage/memstore"
	"github.com/eugene982/yp-gophermart/internal/storage/pgstorage"
)

const (
	// сколько ждём времени на корректное завершение работы сервера
	closeServerTimeout = time.Second * 3
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

// Старт приложения
func run() (err error) {

	// Подготовка логгера
	if err = logger.Initialize("debug"); err != nil {
		return err
	}
	// При выходе логируем ошибку
	defer func() {
		if err != nil {
			logger.Error(err)
		}
	}()

	// захват прерывания процесса
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	conf := config.Config()

	// установка соединения БД
	var store storage.Storage
	if conf.DatabaseDSN == "" {
		store = memstore.New()
	} else {
		var db *sqlx.DB
		db, err = sqlx.Open("pgx", conf.DatabaseDSN)
		if err != nil {
			return
		}
		store, err = pgstorage.New(db)
		if err != nil {
			return
		}
	}

	// клиент, который опрашивает внешний ресурс
	client := &http.Client{
		Timeout: time.Second * time.Duration(conf.Timeout),
	}

	// создание приложения
	app, err := application.New(conf.AccrualSystemAddress, store, client)
	if err != nil {
		return
	}

	server := &http.Server{
		Addr:         conf.ServAddr,
		WriteTimeout: time.Second * time.Duration(conf.Timeout),
		ReadTimeout:  time.Second * time.Duration(conf.Timeout),
		Handler:      app.NewRouter(),
	}

	// запуск сервера в горутине
	srvErr := make(chan error)

	logger.Info("application start", "config", conf)
	go func() {
		srvErr <- server.ListenAndServe()
	}()

	// ждём что раньше случится, ошибка старта сервера
	// или пользователь прервёт программу
	select {
	case <-ctx.Done():
		// прервано пользователем
	case e := <-srvErr:
		// сервер не смог стартануть, некорректый адрес, занят порт...
		// эту ошибку логируем отдельно. В любом случае, нужно освободить ресурсы
		logger.Error(fmt.Errorf("error start server: %w", e))
	}

	// стартуем завершение сервера
	go func() {
		srvErr <- app.Close()
		server.Close() // последним, придёт ErrServerClosed
	}()

	// Ждём пока сервер сам завершится
	// или за отведённое время
	waitClose := time.NewTicker(closeServerTimeout)

	select {
	case <-waitClose.C:
		logger.Info("stop server on timeout")
	case err = <-srvErr:
		logger.Info("stop server gracefull")
	}
	return
}
