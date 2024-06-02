package utils

import (
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
