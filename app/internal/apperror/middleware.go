package apperror

import (
	"errors"
	"net/http"
)

type appHandler func(http.ResponseWriter, *http.Request) error

func Middleware(h appHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var appErr *AppError
		err := h(w, r)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			if errors.As(err, &appErr) {
				//check other custom errors
				if errors.Is(err, ErrNotFound) {
					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write(ErrNotFound.Marshal())
					return
				}

				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write(appErr.Marshal())
				return
			}
			w.WriteHeader(http.StatusTeapot)
			_, _ = w.Write(systemError(err.Error()).Marshal())
		}
	}
}
