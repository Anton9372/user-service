package main

import (
	"Users/internal/app"
	"Users/internal/config"
	"Users/pkg/logging"
	"Users/pkg/shutdown"
	"context"
	"errors"
	"google.golang.org/grpc"
	"net/http"
	"os"
	"syscall"
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logging.InitLogger()
	logger := logging.GetLogger()
	logger.Info("logger initialized")

	logger.Info("config initializing")
	cfg := config.GetConfig()

	a, err := app.NewApp(ctx, cfg, logger)
	if err != nil {
		logger.Fatalf("failed to create application: %v", err)
	}

	go shutdown.Graceful([]os.Signal{syscall.SIGABRT, syscall.SIGQUIT, syscall.SIGHUP, os.Interrupt, syscall.SIGTERM},
		&a)

	logger.Info("start application...")
	err = a.Run(ctx)
	if err != nil && !errors.Is(err, http.ErrServerClosed) && !errors.Is(err, grpc.ErrServerStopped) {
		logger.Fatalf("failed to run application: %v", err)
		return
	}
	logger.Info("gracefully stopped")
}
