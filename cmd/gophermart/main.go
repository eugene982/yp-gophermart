package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/eugene982/yp-gophermart/internal/application"
	"github.com/eugene982/yp-gophermart/internal/config"
	"github.com/eugene982/yp-gophermart/internal/logger"
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
	ctxInterrupt, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	conf := config.Config()
	app, err := application.New(conf)
	if err != nil {
		return
	}

	// запуск сервера в горутине
	srvErr := make(chan error)
	go func() {
		srvErr <- app.Start()
	}()
	logger.Info("application start", "config", conf)

	// ждём что раньше случится, ошибка старта сервера
	// или пользователь прервёт программу
	select {
	case <-ctxInterrupt.Done():
		// прервано пользователем
	case e := <-srvErr:
		// сервер не смог стартануть, некорректый адрес, занят порт...
		// эту ошибку логируем отдельно. В любом случае, нужно освободить ресурсы
		logger.Error(fmt.Errorf("error start server: %w", e))
	}

	// стартуем завершение сервера
	closeErr := make(chan error)
	go func() {
		closeErr <- app.Close()
	}()

	// Ждём пока сервер сам завершится
	// или за отведённое время
	ctxTimeout, stop := context.WithTimeout(context.Background(), closeServerTimeout)
	defer stop()

	select {
	case <-ctxTimeout.Done():
		logger.Warn("stop server on timeout")
	case e := <-closeErr:
		if e != nil {
			err = e
		}
		logger.Info("stop server gracefull")
	}
	return
}
