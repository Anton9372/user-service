package service

import (
	"Users/internal/apperror"
	"Users/internal/user/controller"
	"Users/internal/user/domain/dto"
	"Users/internal/user/domain/model"
	"Users/pkg/logging"
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	Create(ctx context.Context, user model.User) (string, error)
	FindAll(ctx context.Context) ([]model.User, error)
	FindByUUID(ctx context.Context, uuid string) (model.User, error)
	FindByEmail(ctx context.Context, email string) (model.User, error)
	Update(ctx context.Context, user model.User) error
	Delete(ctx context.Context, uuid string) error
}

type service struct {
	repository Repository
	logger     *logging.Logger
}

func NewService(userRepository Repository, logger *logging.Logger) controller.Service {
	return &service{
		repository: userRepository,
		logger:     logger,
	}
}

func (s *service) Create(ctx context.Context, dto dto.CreateUserDTO) (string, error) {
	if dto.Password != dto.RepeatedPassword {
		return "", apperror.BadRequestError("password does not match repeated password")
	}

	user, err := model.NewCreatedUser(dto)
	if err != nil {
		s.logger.Errorf("failed to create user: %v", err)
		return "", err
	}

	var userUUID string
	userUUID, err = s.repository.Create(ctx, user)

	if err != nil {
		s.logger.Errorf("failed to create user: %v", err)
		return userUUID, fmt.Errorf("failed to create user: %w", err)
	}

	return userUUID, nil
}

func (s *service) GetAll(ctx context.Context) ([]model.User, error) {
	users, err := s.repository.FindAll(ctx)

	if err != nil {
		s.logger.Errorf("failed to find all users: %v", err)
		return users, fmt.Errorf("failed to find all users: %w", err)
	}
	return users, nil
}

func (s *service) GetByUUID(ctx context.Context, uuid string) (model.User, error) {
	user, err := s.repository.FindByUUID(ctx, uuid)
	if err != nil {
		s.logger.Errorf("failed to find user by uuid: %v", err)
		if errors.Is(err, apperror.ErrNotFound) {
			return user, err
		}
		return user, fmt.Errorf("failed to find user by uuid. error: %w", err)
	}
	return user, nil
}

func (s *service) GetByEmailAndPassword(ctx context.Context, email, password string) (model.User, error) {
	user, err := s.repository.FindByEmail(ctx, email)
	if err != nil {
		s.logger.Errorf("failed to find user by email: %v", err)
		if errors.Is(err, apperror.ErrNotFound) {
			return user, err
		}
		return user, fmt.Errorf("failed to find user by email: %w", err)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return user, apperror.BadRequestError("incorrect password")
		}
		return user, fmt.Errorf("failed to compare passwords: %w", err)
	}

	return user, nil
}

func (s *service) Update(ctx context.Context, dto dto.UpdateUserDTO) error {
	user, err := s.repository.FindByUUID(ctx, dto.UUID)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(dto.Password))
	if err != nil {
		return apperror.BadRequestError("incorrect password")
	}

	updatedUser, err := model.NewUpdatedUser(user, dto)
	if err != nil {
		return err
	}

	err = s.repository.Update(ctx, updatedUser)

	if err != nil {
		s.logger.Errorf("failed to update user: %v", err)
		return fmt.Errorf("failed to update user: %w", err)
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
		s.logger.Errorf("failed to delete user: %v", err)
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return err
}
