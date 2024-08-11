package controller

import (
	"Users/internal/user/domain/dto"
	"Users/internal/user/domain/model"
	"context"
)

type Service interface {
	Create(ctx context.Context, dto dto.CreateUserDTO) (string, error)
	GetAll(ctx context.Context) ([]model.User, error)
	GetByUUID(ctx context.Context, uuid string) (model.User, error)
	GetByEmailAndPassword(ctx context.Context, email, password string) (model.User, error)
	Update(ctx context.Context, dto dto.UpdateUserDTO) error
	Delete(ctx context.Context, uuid string) error
}
