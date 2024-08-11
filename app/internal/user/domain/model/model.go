package model

import (
	"Users/internal/apperror"
	"Users/internal/user/domain/dto"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	UUID     string `json:"uuid"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewCreatedUser(dto dto.CreateUserDTO) (User, error) {
	user := User{
		Name:     dto.Name,
		Email:    dto.Email,
		Password: dto.Password,
	}
	err := user.GeneratePasswordHash()
	return user, err
}

func NewUpdatedUser(existing User, dto dto.UpdateUserDTO) (User, error) {
	if dto.Name != nil {
		existing.Name = *dto.Name
	}

	if dto.Email != nil {
		existing.Email = *dto.Email
	}

	if dto.NewPassword != nil {
		if dto.RepeatedNewPassword == nil {
			return User{}, apperror.BadRequestError("repeated password must be provided")
		}
		if *dto.NewPassword != *dto.RepeatedNewPassword {
			return User{}, apperror.BadRequestError("passwords do not match")
		}
		existing.Password = *dto.NewPassword
		if err := existing.GeneratePasswordHash(); err != nil {
			return User{}, fmt.Errorf("failed to generate paaword hash: %w", err)
		}
	}
	return existing, nil
}

func (u *User) CheckPassword(password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return fmt.Errorf("password does not match")
	}
	return nil
}

func (u *User) GeneratePasswordHash() error {
	pwdHash, err := generatePasswordHash(u.Password)
	if err != nil {
		return err
	}
	u.Password = pwdHash
	return nil
}

func generatePasswordHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password due to error %w", err)
	}
	return string(hash), nil
}
