package app

import (
	_ "Users/docs"
	"Users/internal/config"
	grpcv1 "Users/internal/user/controller/grpc/v1"
	"Users/internal/user/controller/rest"
	"Users/internal/user/domain/service"
	"Users/internal/user/repository/postgres"
	"Users/pkg/logging"
	"Users/pkg/metric"
	"Users/pkg/postgresql"
	"context"
	"errors"
	"fmt"
	protoUserService "github.com/Anton9372/user-service-contracts/gen/go/user_service/v1"
	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"net/http"
	"time"
)

type App struct {
	cfg               *config.Config
	router            *httprouter.Router
	httpServer        *http.Server
	grpcServer        *grpc.Server
	userServiceServer protoUserService.UserServiceServer

	logger *logging.Logger
}

func NewApp(ctx context.Context, cfg *config.Config, logger *logging.Logger) (App, error) {
	logger.Info("router initializing")
	router := httprouter.New()

	logger.Info("swagger docs initializing")
	router.Handler(http.MethodGet, "/swagger", http.RedirectHandler("/swagger/index.html", http.StatusMovedPermanently))
	router.Handler(http.MethodGet, "/swagger/*any", httpSwagger.WrapHandler)

	logger.Info("http heartbeat initializing")
	metricHandler := metric.Handler{Logger: logger}
	metricHandler.Register(router)

	logger.Info("storage initializing")
	postgresClient, err := postgresql.NewClient(ctx, 5, *cfg)
	if err != nil {
		logger.Fatal(err)
		return App{}, fmt.Errorf("failed to init storage: %w", err)
	}

	userStorage := postgres.NewRepository(postgresClient, logger)
	userService := service.NewService(userStorage, logger)

	usersHandler := rest.NewHandler(userService, logger)
	usersHandler.Register(router)

	usersGRPCServer := grpcv1.NewServer(protoUserService.UnimplementedUserServiceServer{}, userService, logger)

	return App{
		cfg:               cfg,
		router:            router,
		userServiceServer: usersGRPCServer,
		logger:            logger,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		return a.startHTTP()
	})

	group.Go(func() error {
		return a.startGRPC(a.userServiceServer)
	})

	return group.Wait()
}

func (a *App) startGRPC(server protoUserService.UserServiceServer) error {
	a.logger.Info("gRPC server initializing")
	a.logger.Infof("bind gRPC to host: %s and port: %d", a.cfg.GRPC.IP, a.cfg.GRPC.Port)

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", a.cfg.GRPC.IP, a.cfg.GRPC.Port))
	if err != nil {
		a.logger.Fatalf("failed to create listener: %v", err)
	}

	var serverOptions []grpc.ServerOption
	a.grpcServer = grpc.NewServer(serverOptions...)
	protoUserService.RegisterUserServiceServer(a.grpcServer, server)
	reflection.Register(a.grpcServer)

	a.logger.Info("gRPC server started")

	if err = a.grpcServer.Serve(listener); err != nil {
		switch {
		case errors.Is(err, grpc.ErrServerStopped):
			a.logger.Warn("gRPC server shutdown")
		default:
			a.logger.Fatalf("failed to start server: %v", err)
		}
	}

	return err
}

func (a *App) startHTTP() error {
	a.logger.Info("HTTP server initializing")
	a.logger.Infof("bind HTTP to host: %s and port: %d", a.cfg.HTTP.IP, a.cfg.HTTP.Port)

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", a.cfg.HTTP.IP, a.cfg.HTTP.Port))
	if err != nil {
		a.logger.Fatalf("failed to create listener: %v", err)
	}

	c := cors.New(cors.Options{
		AllowedMethods:   a.cfg.HTTP.CORS.AllowedMethods,
		AllowedOrigins:   a.cfg.HTTP.CORS.AllowedOrigins,
		AllowCredentials: a.cfg.HTTP.CORS.AllowCredentials,
		AllowedHeaders:   a.cfg.HTTP.CORS.AllowedHeaders,
		ExposedHeaders:   a.cfg.HTTP.CORS.ExposedHeaders,
	})

	handler := c.Handler(a.router)

	a.httpServer = &http.Server{
		Handler: handler,
		//TODO to config
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	a.logger.Info("HTTP server started")

	if err = a.httpServer.Serve(listener); err != nil {
		switch {
		case errors.Is(err, http.ErrServerClosed):
			a.logger.Warn("HTTP server shutdown")
		default:
			a.logger.Fatalf("failed to start server: %v", err)
		}
	}

	return err
}

func (a *App) Close() error {
	err := a.httpServer.Close()
	if err != nil {
		a.logger.Errorf("failed to close HTTP server: %v", err)
	}

	a.grpcServer.GracefulStop()

	return err
}
