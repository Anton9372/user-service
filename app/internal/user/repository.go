package user

import (
	"context"
)

type Repository interface {
	Create(ctx context.Context, user User) (string, error)
	FindAll(ctx context.Context) ([]User, error)
	FindByUUID(ctx context.Context, uuid string) (User, error)
	FindByEmail(ctx context.Context, email string) (User, error)
	Update(ctx context.Context, user User) error
	Delete(ctx context.Context, uuid string) error
}
