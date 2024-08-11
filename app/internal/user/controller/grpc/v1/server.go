package grpc

import (
	"Users/internal/user/controller"
	"Users/pkg/logging"
	"context"
	protoUserService "github.com/Anton9372/user-service-contracts/gen/go/user_service/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	protoUserService.UnimplementedUserServiceServer
	service controller.Service
	logger  *logging.Logger
}

func NewServer(
	protoService protoUserService.UnimplementedUserServiceServer,
	userService controller.Service,
	logger *logging.Logger,
) *Server {
	return &Server{
		UnimplementedUserServiceServer: protoService,
		service:                        userService,
		logger:                         logger,
	}
}

func (s *Server) Create(
	ctx context.Context, req *protoUserService.CreateRequest,
) (*protoUserService.CreateResponse, error) {
	s.logger.Debug("Create user")
	input, err := NewCreateUserDTO(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	output, err := s.service.Create(ctx, input)
	if err != nil {
		return nil, HandleServiceError(err)
	}

	return &protoUserService.CreateResponse{Uuid: output}, nil
}

func (s *Server) GetByUUID(
	ctx context.Context, req *protoUserService.GetByUUIDRequest,
) (*protoUserService.UserResponse, error) {
	s.logger.Debug("Get user by uuid")

	if req.Uuid == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user's uuid is required")
	}

	user, err := s.service.GetByUUID(ctx, req.Uuid)
	if err != nil {
		return nil, HandleServiceError(err)
	}

	return &protoUserService.UserResponse{User: NewProtoUser(user)}, nil
}

func (s *Server) GetByEmailAndPassword(
	ctx context.Context, req *protoUserService.GetByEmailAndPasswordRequest,
) (*protoUserService.UserResponse, error) {
	s.logger.Debug("Get user by email and password")

	if req.Email == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email must not be empty")
	}
	if req.Password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "password must not be empty")
	}

	user, err := s.service.GetByEmailAndPassword(ctx, req.Email, req.Password)
	if err != nil {
		return nil, HandleServiceError(err)
	}

	return &protoUserService.UserResponse{User: NewProtoUser(user)}, nil
}

func (s *Server) Update(
	ctx context.Context, req *protoUserService.UpdateRequest,
) (*protoUserService.UpdateResponse, error) {
	s.logger.Debug("Partially update user")
	input, err := NewUpdateUserDTO(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	err = s.service.Update(ctx, input)
	if err != nil {
		return nil, HandleServiceError(err)
	}

	return &protoUserService.UpdateResponse{}, nil
}

func (s *Server) Delete(
	ctx context.Context, req *protoUserService.DeleteRequest,
) (*protoUserService.DeleteResponse, error) {
	s.logger.Debug("Delete user")
	if req.Uuid == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user's uuid must not be empty")
	}

	err := s.service.Delete(ctx, req.Uuid)
	if err != nil {
		return nil, HandleServiceError(err)
	}

	return &protoUserService.DeleteResponse{}, nil
}
