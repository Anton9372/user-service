package grpc

import (
	"Users/internal/apperror"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func HandleServiceError(err error) error {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		//check other custom errors
		if errors.Is(err, apperror.ErrNotFound) {
			return status.Error(codes.NotFound, err.Error())
		}

		return status.Error(codes.InvalidArgument, err.Error())
	}
	return status.Error(codes.Internal, err.Error())
}
