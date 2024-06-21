package utils

import (
	"Users/pkg/logging"
	"io"
	"time"
)

func DoWithAttempts(fn func() error, attempts int, delay time.Duration) error {
	var err error
	for attempts > 0 {
		if err = fn(); err != nil {
			time.Sleep(delay)
			attempts--
			continue
		}
		return nil
	}
	return err
}

func CloseBody(logger *logging.Logger, body io.ReadCloser) {
	err := body.Close()
	if err != nil {
		logger.Fatalf("Error closing request body: %v", err)
	}
}
