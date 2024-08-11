package dto

import "fmt"

type CreateUserDTO struct {
	Name             string `json:"name"`
	Email            string `json:"email"`
	Password         string `json:"password"`
	RepeatedPassword string `json:"repeated_password"`
}

func (dto *CreateUserDTO) ValidateEmptyFields() error {
	if dto.Name == "" {
		return fmt.Errorf("name must not be empty")
	}
	if dto.Email == "" {
		return fmt.Errorf("email must not be empty")
	}
	if dto.Password == "" {
		return fmt.Errorf("password must not be empty")
	}
	if dto.RepeatedPassword == "" {
		return fmt.Errorf("repeated password must not be empty")
	}
	return nil
}

type UpdateUserDTO struct {
	UUID                string  `json:"uuid,omitempty"`
	Name                *string `json:"name,omitempty"`
	Email               *string `json:"email,omitempty"`
	Password            string  `json:"password,omitempty"`
	NewPassword         *string `json:"new_password,omitempty"`
	RepeatedNewPassword *string `json:"repeated_new_password,omitempty"`
}

func (dto *UpdateUserDTO) ValidateEmptyFields() error {
	if dto.UUID == "" {
		return fmt.Errorf("uuid must not be empty")
	}
	if dto.Password == "" {
		return fmt.Errorf("password must not be empty")
	}
	return nil
}
