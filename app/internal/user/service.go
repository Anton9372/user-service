package user

import (
	"Users/internal/apperror"
	"Users/pkg/logging"
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	Create(ctx context.Context, user User) (string, error)
	FindAll(ctx context.Context) ([]User, error)
	FindByUUID(ctx context.Context, uuid string) (User, error)
	FindByEmail(ctx context.Context, email string) (User, error)
	Update(ctx context.Context, user User) error
	Delete(ctx context.Context, uuid string) error
}

type service struct {
	repository Repository
	logger     *logging.Logger
}

func NewService(userRepository Repository, logger *logging.Logger) Service {
	return &service{
		repository: userRepository,
		logger:     logger,
	}
}

func (s *service) Create(ctx context.Context, dto CreateUserDTO) (string, error) {
	if dto.Password != dto.RepeatedPassword {
		return "", apperror.BadRequestError("password does not match repeated password")
	}

	user := NewUser(dto)

	err := user.GeneratePasswordHash()
	if err != nil {
		s.logger.Fatalf("failed to create user. error: %v", err)
		return "", err
	}

	var userUUID string
	userUUID, err = s.repository.Create(ctx, user)

	if err != nil {
		return userUUID, fmt.Errorf("failed to create user. error: %w", err)
	}

	return userUUID, nil
}

func (s *service) GetAll(ctx context.Context) ([]User, error) {
	users, err := s.repository.FindAll(ctx)

	if err != nil {
		return users, fmt.Errorf("failed to find all users: %w", err)
	}
	return users, nil
}

func (s *service) GetByUUID(ctx context.Context, uuid string) (User, error) {
	user, err := s.repository.FindByUUID(ctx, uuid)

	if err != nil {
		return user, fmt.Errorf("failed to find user by uuid. error: %w", err)
	}
	return user, nil
}

func (s *service) GetByEmailAndPassword(ctx context.Context, email, password string) (User, error) {
	user, err := s.repository.FindByEmail(ctx, email)

	if err != nil {
		return user, fmt.Errorf("failed to find user by email. error: %w", err)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return user, apperror.BadRequestError("incorrect password")
		}
		return user, fmt.Errorf("failed to compare passwords. error: %w", err)
	}

	return user, nil
}

func (s *service) Update(ctx context.Context, dto UpdateUserDTO) error {
	if dto.NewPassword != dto.RepeatedNewPassword {
		return apperror.BadRequestError("password does not match repeated password")
	}
	user, err := s.repository.FindByUUID(ctx, dto.UUID)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(dto.OldPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return apperror.BadRequestError("incorrect password")
		}
		return apperror.BadRequestError("old password does not match current password")
	}

	updatedUser, err := UpdatedUser(user, dto)
	if err != nil {
		return err
	}

	err = s.repository.Update(ctx, *updatedUser)

	if err != nil {
		return fmt.Errorf("failed to update user. error: %w", err)
	}
	return nil
}

func (s *service) Delete(ctx context.Context, uuid string) error {
	_, err := s.repository.FindByUUID(ctx, uuid)
	if err != nil {
		return err
	}

	err = s.repository.Delete(ctx, uuid)

	if err != nil {
		return fmt.Errorf("failed to delete user. error: %w", err)
	}
	return err
}
