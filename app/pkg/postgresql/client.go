package postgresql

import (
	"Users/internal/config"
	"Users/pkg/utils"
	"context"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
	"time"
)

const waitTime = 5 * time.Second

type Client interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
}

func NewClient(ctx context.Context, connectionAttempts int, cfg config.Config) (*pgxpool.Pool, error) {
	var dsn string
	if cfg.Postgres.Username == "" && cfg.Postgres.Password == "" {
		dsn = fmt.Sprintf("postgresql://%s:%s/%s",
			cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.Database)
	} else {
		dsn = fmt.Sprintf("postgresql://%s:%s@%s:%s/%s",
			cfg.Postgres.Username, cfg.Postgres.Password, cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.Database)
	}

	var pool *pgxpool.Pool
	err := utils.DoWithAttempts(func() error {
		ctx, cancel := context.WithTimeout(ctx, waitTime)
		defer cancel()

		var pgxErr error
		pool, pgxErr = pgxpool.Connect(ctx, dsn)
		if pgxErr != nil {
			return pgxErr
		}

		return nil
	}, connectionAttempts, waitTime)

	if err != nil {
		logrus.Fatal("Error connecting to database: ", err)
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	err = pool.Ping(context.Background())
	if err != nil {
		logrus.Fatal("Error pinging database: ", err)
		return nil, fmt.Errorf("error pinging database: %v", err)
	}

	return pool, nil
}
