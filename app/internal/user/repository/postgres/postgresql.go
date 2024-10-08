package postgres

import (
	"Users/internal/apperror"
	"Users/internal/user/domain/model"
	"Users/internal/user/domain/service"
	"Users/pkg/logging"
	"Users/pkg/postgresql"
	"Users/pkg/utils"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"time"
)

const queryWaitTime = 5 * time.Second

type repository struct {
	client postgresql.Client
	logger *logging.Logger
}

func NewRepository(client postgresql.Client, logger *logging.Logger) service.Repository {
	return &repository{
		client: client,
		logger: logger,
	}
}

func handleSQLError(err error, logger *logging.Logger) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return apperror.ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		newErr := fmt.Errorf("SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s",
			pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState())
		logger.Error(newErr)

		if pgErr.Code == "23505" { //uniqueness violation
			return apperror.BadRequestError("User with this email already exists")
		} else if pgErr.Code == "22P02" { //invalid uuid syntax
			return apperror.ErrNotFound
		}
		return newErr
	}

	return err
}

func (r *repository) Create(ctx context.Context, user model.User) (string, error) {
	query := `
				INSERT INTO users
					(name, email, password)
				VALUES
					($1, $2, $3)
				RETURNING id;
	`
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatSQLQuery(query)))

	var userUUID string
	nCtx, cancel := context.WithTimeout(ctx, queryWaitTime)
	defer cancel()
	err := r.client.QueryRow(nCtx, query, user.Name, user.Email, user.Password).Scan(&userUUID)
	if err != nil {
		return "", handleSQLError(err, r.logger)
	}

	return userUUID, nil
}

func (r *repository) FindAll(ctx context.Context) ([]model.User, error) {
	query := `
				SELECT
					id, name, email, password
				FROM
					users
	`
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatSQLQuery(query)))

	nCtx, cancel := context.WithTimeout(ctx, queryWaitTime)
	defer cancel()
	rows, err := r.client.Query(nCtx, query)
	if err != nil {
		return nil, handleSQLError(err, r.logger)
	}
	defer rows.Close()

	users := make([]model.User, 0)
	for rows.Next() {
		var usr model.User
		err = rows.Scan(&usr.UUID, &usr.Name, &usr.Email, &usr.Password)
		if err != nil {
			return nil, err
		}
		users = append(users, usr)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *repository) FindByUUID(ctx context.Context, uuid string) (model.User, error) {
	query := `
				SELECT
					id, name, email, password
				FROM
					users
				WHERE
					id = $1
	`
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatSQLQuery(query)))

	var usr model.User
	nCtx, cancel := context.WithTimeout(ctx, queryWaitTime)
	defer cancel()
	err := r.client.QueryRow(nCtx, query, uuid).Scan(&usr.UUID, &usr.Name, &usr.Email, &usr.Password)
	if err != nil {
		return model.User{}, handleSQLError(err, r.logger)
	}
	return usr, nil
}

func (r *repository) FindByEmail(ctx context.Context, email string) (model.User, error) {
	query := `
				SELECT
					id, name, email, password
				FROM
					users
				WHERE
					email = $1
    `
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatSQLQuery(query)))

	var usr model.User
	nCtx, cancel := context.WithTimeout(ctx, queryWaitTime)
	defer cancel()
	err := r.client.QueryRow(nCtx, query, email).Scan(&usr.UUID, &usr.Name, &usr.Email, &usr.Password)
	if err != nil {
		return model.User{}, handleSQLError(err, r.logger)
	}
	return usr, nil
}

func (r *repository) Update(ctx context.Context, user model.User) error {
	query := `
				UPDATE
					users
				SET
					name = $1, email = $2, password = $3
				WHERE
					id = $4
    `
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatSQLQuery(query)))

	nCtx, cancel := context.WithTimeout(ctx, queryWaitTime)
	defer cancel()
	cmdTag, err := r.client.Exec(nCtx, query, user.Name, user.Email, user.Password, user.UUID)
	if err != nil {
		return handleSQLError(err, r.logger)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("no rows were updated")
	}

	return nil
}

func (r *repository) Delete(ctx context.Context, uuid string) error {
	query := `
				DELETE
				FROM
					users
				WHERE
					id = $1
    `
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatSQLQuery(query)))

	nCtx, cancel := context.WithTimeout(ctx, queryWaitTime)
	defer cancel()
	cmdTag, err := r.client.Exec(nCtx, query, uuid)
	if err != nil {
		return handleSQLError(err, r.logger)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("no rows were deleted")
	}

	return nil
}
