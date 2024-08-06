package model

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	UUID     string `json:"uuid"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewUser(dto CreateUserDTO) User {
	return User{
		Name:     dto.Name,
		Email:    dto.Email,
		Password: dto.Password,
	}
}

func UpdatedUser(existing User, dto UpdateUserDTO) (*User, error) {
	updUser := new(User)

	updUser.UUID = dto.UUID

	if dto.Name != "" {
		updUser.Name = dto.Name
	} else {
		updUser.Name = existing.Name
	}

	if dto.Email != "" {
		updUser.Email = dto.Email
	} else {
		updUser.Email = existing.Email
	}

	if dto.NewPassword != "" {
		updUser.Password = dto.NewPassword
		err := updUser.GeneratePasswordHash()
		if err != nil {
			return &User{}, fmt.Errorf("failed to create updated user: %w", err)
		}
	} else {
		updUser.Password = existing.Password
	}
	return updUser, nil
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
