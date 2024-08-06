package main

import (
	_ "Users/docs"
	"Users/internal/config"
	"Users/internal/user/controller/rest"
	"Users/internal/user/domain/service"
	"Users/internal/user/repository/postgres"
	"Users/pkg/logging"
	"Users/pkg/metric"
	"Users/pkg/postgresql"
	"Users/pkg/shutdown"
	"context"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"
)

// @Title		User-service API
// @Version		1.0
// @Description	Service for user management

// @Contact.name	Anton
// @Contact.email	ap363402@gmail.com

// @License.name Apache 2.0

// @Host 		localhost:10001
// @BasePath 	/api
func main() {
	logging.InitLogger()
	logger := logging.GetLogger()
	logger.Info("logger initialized")

	logger.Info("config initializing")
	cfg := config.GetConfig()

	logger.Info("router initializing")
	router := httprouter.New()

	logger.Info("swagger docs initializing")
	router.Handler(http.MethodGet, "/swagger", http.RedirectHandler("/swagger/index.html", http.StatusMovedPermanently))
	router.Handler(http.MethodGet, "/swagger/*any", httpSwagger.WrapHandler)

	metricHandler := metric.Handler{Logger: logger}
	metricHandler.Register(router)

	logger.Info("storage initializing")
	postgresClient, err := postgresql.NewClient(context.Background(), 5, *cfg)
	if err != nil {
		logger.Fatal(err)
	}
	userStorage := postgres.NewRepository(postgresClient, logger)
	userService := service.NewService(userStorage, logger)
	usersHandler := rest.NewHandler(userService, logger)
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
