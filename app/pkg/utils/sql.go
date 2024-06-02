package utils

import (
	"Users/pkg/logging"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"strings"
)

func FormatSQLQuery(query string) string {
	return strings.ReplaceAll(strings.ReplaceAll(query, "\t", ""), "\n", " ")
}

func HandleSQLError(err error, logger *logging.Logger) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		newErr := fmt.Errorf("SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s",
			pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState())
		logger.Error(newErr)
		return newErr
	}
	return err
}
