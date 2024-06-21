package user

import (
	"Users/internal/apperror"
	h "Users/internal/handler"
	"Users/pkg/logging"
	"Users/pkg/utils"
	"context"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

const (
	usersURL    = "/api/users"
	userByIdURL = "/api/users/one/:uuid"
	allUsersURL = "/api/users/all"
)

type Service interface {
	Create(ctx context.Context, dto CreateUserDTO) (string, error)
	GetAll(ctx context.Context) ([]User, error)
	GetByUUID(ctx context.Context, uuid string) (User, error)
	GetByEmailAndPassword(ctx context.Context, dto EmailAndPasswordDTO) (User, error)
	Update(ctx context.Context, dto UpdateUserDTO) error
	Delete(ctx context.Context, uuid string) error
}

type handler struct {
	service Service
	logger  *logging.Logger
}

func NewHandler(service Service, logger *logging.Logger) h.Handler {
	return &handler{
		service: service,
		logger:  logger,
	}
}

func (h *handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodPost, usersURL, apperror.Middleware(h.CreateUser))
	router.HandlerFunc(http.MethodGet, allUsersURL, apperror.Middleware(h.GetAllUsers))
	router.HandlerFunc(http.MethodGet, userByIdURL, apperror.Middleware(h.GetUserByUUID))
	router.HandlerFunc(http.MethodGet, usersURL, apperror.Middleware(h.GetUserByEmailAndPassword))
	router.HandlerFunc(http.MethodPatch, userByIdURL, apperror.Middleware(h.PartiallyUpdateUser))
	router.HandlerFunc(http.MethodDelete, userByIdURL, apperror.Middleware(h.DeleteUser))
}

func (h *handler) CreateUser(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("Create user")
	defer utils.CloseBody(h.logger, r.Body)
	w.Header().Set("Content-Type", "application/json")

	var createdUser CreateUserDTO

	if err := json.NewDecoder(r.Body).Decode(&createdUser); err != nil {
		return apperror.BadRequestError("invalid JSON scheme. check swagger API")
	}

	if createdUser.Name == "" || createdUser.Email == "" || createdUser.Password == "" ||
		createdUser.RepeatedPassword == "" {
		return apperror.BadRequestError("missing required fields")
	}

	userUUID, err := h.service.Create(r.Context(), createdUser)
	if err != nil {
		return err
	}
	w.Header().Set("Location", fmt.Sprintf("%s/%s", usersURL, userUUID))
	w.WriteHeader(http.StatusCreated)

	h.logger.Info("Create user successfully")
	return nil
}

func (h *handler) GetAllUsers(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("Get all users")
	defer utils.CloseBody(h.logger, r.Body)
	w.Header().Set("Content-Type", "application/json")

	users, err := h.service.GetAll(r.Context())
	if err != nil {
		return err
	}

	userBytes, err := json.Marshal(users)
	if err != nil {
		return fmt.Errorf("failed to marshall users %w", err)
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(userBytes)
	if err != nil {
		return err
	}

	h.logger.Info("Get all users successfully")
	return nil
}

func (h *handler) GetUserByUUID(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("Get user by uuid")
	defer utils.CloseBody(h.logger, r.Body)
	w.Header().Set("Content-Type", "application/json")

	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	userUUID := params.ByName("uuid")
	if userUUID == "" {
		return apperror.BadRequestError("user uuid must not be empty")
	}

	usr, err := h.service.GetByUUID(r.Context(), userUUID)
	if err != nil {
		return err
	}

	userBytes, err := json.Marshal(usr)
	if err != nil {
		return fmt.Errorf("failed to marshall user. error: %w", err)
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(userBytes)
	if err != nil {
		return err
	}

	h.logger.Info("Get user by uuid successfully")
	return nil
}

func (h *handler) GetUserByEmailAndPassword(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("Get user by email and password")
	defer utils.CloseBody(h.logger, r.Body)
	w.Header().Set("Content-Type", "application/json")

	var emailAndPassword EmailAndPasswordDTO

	if err := json.NewDecoder(r.Body).Decode(&emailAndPassword); err != nil {
		return apperror.BadRequestError("invalid JSON scheme. check swagger API")
	}

	if emailAndPassword.Email == "" || emailAndPassword.Password == "" {
		return apperror.BadRequestError("missing required fields")
	}

	user, err := h.service.GetByEmailAndPassword(r.Context(), emailAndPassword)
	if err != nil {
		return err
	}

	userBytes, err := json.Marshal(user)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(userBytes)
	if err != nil {
		return err
	}

	h.logger.Info("Get user by email and password successfully")
	return nil
}

func (h *handler) PartiallyUpdateUser(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("Partially update user")
	defer utils.CloseBody(h.logger, r.Body)
	w.Header().Set("Content-Type", "application/json")

	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	userUUID := params.ByName("uuid")
	if userUUID == "" {
		return apperror.BadRequestError("user uuid must not be empty")
	}

	var updatedUser UpdateUserDTO

	if err := json.NewDecoder(r.Body).Decode(&updatedUser); err != nil {
		return apperror.BadRequestError("invalid JSON scheme. check swagger API")
	}

	updatedUser.UUID = userUUID

	err := h.service.Update(r.Context(), updatedUser)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)

	h.logger.Info("Partially update user successfully")
	return nil
}

func (h *handler) DeleteUser(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("Delete user")
	defer utils.CloseBody(h.logger, r.Body)
	w.Header().Set("Content-Type", "application/json")

	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	userUUID := params.ByName("uuid")
	if userUUID == "" {
		return apperror.BadRequestError("user uuid must not be empty")
	}

	err := h.service.Delete(r.Context(), userUUID)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)

	h.logger.Info("Delete user successfully")
	return nil
}
