package user

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

func UpdatedUser(dto UpdateUserDTO) User {
	return User{
		UUID:     dto.UUID,
		Name:     dto.Name,
		Email:    dto.Email,
		Password: dto.NewPassword,
	}
}

func (u *User) MergeWithDefaults(defaultUser User) error {
	if u.UUID == "" {
		u.UUID = defaultUser.UUID
	}
	if u.Name == "" {
		u.Name = defaultUser.Name
	}
	if u.Email == "" {
		u.Email = defaultUser.Email
	}
	if u.Password == "" {
		u.Password = defaultUser.Password
	} else {
		err := u.GeneratePasswordHash()
		if err != nil {
			return fmt.Errorf("failed to update user. error %w", err)
		}
	}
	return nil
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
