package utils

import (
	"errors"
	"github.com/jackc/pgconn"
	"log/slog"
	"os"
)

type ErrorMessage struct {
	ErrorMessage string `json:"error_message"`
}

type SuccessMessage struct {
	SuccessMessage string `json:"success_message"`
}

func LogErrorFatal(err error) {
	slog.Error(err.Error())
	os.Exit(1)
}

func IsStorageRelatedError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return true
	}
	return false
}
