package db

import (
	"Users/internal/user"
	"Users/pkg/logging"
	"Users/pkg/postgresql"
	"Users/pkg/utils"
	"context"
	"fmt"
	"time"
)

const queryWaitTime = 5 * time.Second

type repository struct {
	client postgresql.Client
	logger *logging.Logger
}

func NewRepository(client postgresql.Client, logger *logging.Logger) user.Repository {
	return &repository{
		client: client,
		logger: logger,
	}
}

func (r *repository) Create(ctx context.Context, user user.User) (string, error) {
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
		return "", utils.HandleSQLError(err, r.logger)
	}

	return userUUID, nil
}

func (r *repository) FindAll(ctx context.Context) ([]user.User, error) {
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
		return nil, utils.HandleSQLError(err, r.logger)
	}
	defer rows.Close()

	users := make([]user.User, 0)
	for rows.Next() {
		var usr user.User
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

func (r *repository) FindByUUID(ctx context.Context, uuid string) (user.User, error) {
	query := `
				SELECT
					id, name, email, password
				FROM
					users
				WHERE
				    id = $1
	`
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatSQLQuery(query)))

	var usr user.User
	nCtx, cancel := context.WithTimeout(ctx, queryWaitTime)
	defer cancel()
	err := r.client.QueryRow(nCtx, query, uuid).Scan(&usr.UUID, &usr.Name, &usr.Email, &usr.Password)
	if err != nil {
		return user.User{}, utils.HandleSQLError(err, r.logger)
	}
	return usr, nil
}

func (r *repository) FindByEmail(ctx context.Context, email string) (user.User, error) {
	query := `
        		SELECT
            		id, name, email, password
        		FROM
            		users
       			WHERE
            		email = $1
    `
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatSQLQuery(query)))

	var usr user.User
	nCtx, cancel := context.WithTimeout(ctx, queryWaitTime)
	defer cancel()
	err := r.client.QueryRow(nCtx, query, email).Scan(&usr.UUID, &usr.Name, &usr.Email, &usr.Password)
	if err != nil {
		return user.User{}, utils.HandleSQLError(err, r.logger)
	}
	return usr, nil
}

func (r *repository) Update(ctx context.Context, user user.User) error {
	query := `
        		UPDATE 
        		    users
        		SET 
        		    name = $1, email = $2, password = $3
        		WHERE 
        		    id = $4
    `
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatSQLQuery(query)))

	cmdTag, err := r.client.Exec(ctx, query, user.Name, user.Email, user.Password, user.UUID)
	if err != nil {
		return utils.HandleSQLError(err, r.logger)
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

	cmdTag, err := r.client.Exec(ctx, query, uuid)
	if err != nil {
		return utils.HandleSQLError(err, r.logger)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("no rows were deleted")
	}

	return nil
}
