package grpc

import (
	"Users/internal/user/domain/dto"
	"Users/internal/user/domain/model"
	protoUserService "github.com/Anton9372/user-service-contracts/gen/go/user_service/v1"
)

func NewProtoUser(user model.User) *protoUserService.User {
	return &protoUserService.User{
		Uuid:     user.UUID,
		Name:     user.Name,
		Email:    user.Email,
		Password: user.Password,
	}
}

func NewCreateUserDTO(req *protoUserService.CreateRequest) (dto.CreateUserDTO, error) {
	createdUser := dto.CreateUserDTO{
		Name:             req.Name,
		Email:            req.Email,
		Password:         req.Password,
		RepeatedPassword: req.RepeatedPassword,
	}
	err := createdUser.ValidateEmptyFields()
	return createdUser, err
}

func NewUpdateUserDTO(req *protoUserService.UpdateRequest) (dto.UpdateUserDTO, error) {
	updatedUser := dto.UpdateUserDTO{
		UUID:                req.Uuid,
		Name:                req.Name,
		Email:               req.Email,
		Password:            req.Password,
		NewPassword:         req.NewPassword,
		RepeatedNewPassword: req.RepeatedNewPassword,
	}

	err := updatedUser.ValidateEmptyFields()
	return updatedUser, err
}
