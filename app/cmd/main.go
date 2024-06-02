package main

import (
	"Users/internal/config"
	"Users/internal/user"
	"Users/internal/user/db"
	"Users/pkg/logging"
	"Users/pkg/metric"
	"Users/pkg/postgresql"
	"Users/pkg/shutdown"
	"context"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"
)

func main() {
	logging.InitLogger()
	logger := logging.GetLogger()
	logger.Info("logger initialized")

	logger.Info("config initializing")
	cfg := config.GetConfig()

	logger.Info("router initializing")
	router := httprouter.New()

	metricHandler := metric.Handler{Logger: logger}
	metricHandler.Register(router)

	postgresClient, err := postgresql.NewClient(context.Background(), 5, *cfg)
	if err != nil {
		logger.Fatal(err)
	}
	userStorage := db.NewRepository(postgresClient, logger)
	userService := user.NewService(userStorage, logger)
	usersHandler := user.NewHandler(userService, logger)
	usersHandler.Register(router)

	logger.Info("start application")
	start(router, logger, cfg)
}

func start(router http.Handler, logger *logging.Logger, cfg *config.Config) {
	var server *http.Server
	var listener net.Listener
	var err error

	logger.Infof("bind application to host: %s and port: %s", cfg.Listen.BindIP, cfg.Listen.Port)

	listener, err = net.Listen("tcp", fmt.Sprintf("%s:%s", cfg.Listen.BindIP, cfg.Listen.Port))
	if err != nil {
		logger.Fatal(err)
	}

	server = &http.Server{
		Handler:      router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go shutdown.Graceful([]os.Signal{syscall.SIGABRT, syscall.SIGQUIT, syscall.SIGHUP, os.Interrupt, syscall.SIGTERM},
		server)

	logger.Info("application initialized and started")

	if err := server.Serve(listener); err != nil {
		switch {
		case errors.Is(err, http.ErrServerClosed):
			logger.Warn("server shutdown")
		default:
			logger.Fatal(err)
		}
	}
}
