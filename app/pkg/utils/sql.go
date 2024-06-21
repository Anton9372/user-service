package utils

import (
	"strings"
)

func FormatSQLQuery(query string) string {
	return strings.ReplaceAll(strings.ReplaceAll(query, "\t", ""), "\n", " ")
}
