package main

import (
	"context"
	echoSwagger "github.com/swaggo/echo-swagger"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"identity-forecaster/internal/app/forecaster/config"
	"identity-forecaster/internal/app/forecaster/handler"
	"identity-forecaster/internal/app/forecaster/repository"
	"identity-forecaster/internal/app/forecaster/service"
	"identity-forecaster/internal/pkg/logger"

	_ "identity-forecaster/docs"
)

func router(pgPool *pgxpool.Pool, wg *sync.WaitGroup, retriesAmount uint, msBetweenRetries uint, APIs []string) (*echo.Echo, error) {
	e := echo.New()

	pg := repository.NewPostgres(pgPool)

	r := repository.New(pg)
	s := service.New(r)
	h := handler.New(s, wg, retriesAmount, msBetweenRetries, APIs)

	e.POST("/create", h.CreatePerson)
	e.DELETE("/delete/:id", h.DeletePersonByID)
	e.PUT("/update/:id", h.UpdatePerson)
	e.GET("/read", h.ReadPersons)
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	return e, nil
}

// @title Identity Forecaster API
// @version 1.0
// @description Сервис, получающий ФИО, и обогащающий информацию о нем из открытых источников

// @host localhost:8787
// @BasePath /

// @Tag.name Persons
// @Tag.description Группа запросов для управления сущностями

// @Schemes http

func main() {
	cfg := config.LoadConfig()

	logger.SetLogfilePath(cfg.Logfile)
	logger.Logger().Infoln("config loaded")

	pgPool, err := repository.GetPgxPool(cfg.DatabaseDSN)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup

	r, err := router(pgPool, &wg, cfg.RetriesAmount, cfg.RetryIntervalMilliseconds, cfg.APIs)
	if err != nil {
		panic(err)
	}

	go func() {
		err = r.Start(cfg.ServiceHost + ":" + cfg.ServicePort)
		if err != nil {
			logger.Logger().Infoln(err)
		}
	}()

	c := make(chan os.Signal, 1)
	ret := make(chan struct{}, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM)
	go func() {
		<-c
		ret <- struct{}{}
	}()

	<-ret
	logger.Logger().Infoln("shutting down gracefully...")
	const timeoutInterval = 5 * time.Second
	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeoutInterval)

	if err = r.Shutdown(shutdownCtx); err != nil {
		logger.Logger().Infoln(err)
		return
	} else {
		cancel()
	}

	<-shutdownCtx.Done()

	wg.Wait()
}
