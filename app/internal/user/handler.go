package user

import (
	"Users/internal/apperror"
	"Users/pkg/logging"
	"context"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io"
	"net/http"
)

const (
	usersURL                  = "/api/users"
	userByIdURL               = "/api/users/uuid/:uuid"
	userByEmailAndPasswordURL = "/api/users/login"
)

type Service interface {
	Create(ctx context.Context, dto CreateUserDTO) (string, error)
	GetAll(ctx context.Context) ([]User, error)
	GetByUUID(ctx context.Context, uuid string) (User, error)
	GetByEmailAndPassword(ctx context.Context, email, password string) (User, error)
	Update(ctx context.Context, dto UpdateUserDTO) error
	Delete(ctx context.Context, uuid string) error
}

type Handler struct {
	service Service
	logger  *logging.Logger
}

func NewHandler(service Service, logger *logging.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodPost, usersURL, apperror.Middleware(h.CreateUser))
	router.HandlerFunc(http.MethodGet, usersURL, apperror.Middleware(h.GetAllUsers))
	router.HandlerFunc(http.MethodGet, userByIdURL, apperror.Middleware(h.GetUserByUUID))
	router.HandlerFunc(http.MethodGet, userByEmailAndPasswordURL, apperror.Middleware(h.GetUserByEmailAndPassword))
	router.HandlerFunc(http.MethodPatch, userByIdURL, apperror.Middleware(h.PartiallyUpdateUser))
	router.HandlerFunc(http.MethodDelete, userByIdURL, apperror.Middleware(h.DeleteUser))
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("Create user")
	w.Header().Set("Content-Type", "application/json")

	var createdUser CreateUserDTO
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			h.logger.Errorf("Error closing body %v", err)
		}
	}(r.Body)
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

	return nil
}

func (h *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("Get all users")
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
	return nil
}

func (h *Handler) GetUserByUUID(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("Get user by uuid")
	w.Header().Set("Content-Type", "application/json")

	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	userUUID := params.ByName("uuid")

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
	return nil
}

func (h *Handler) GetUserByEmailAndPassword(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("Get user by email and password")
	w.Header().Set("Content-Type", "application/json")

	email := r.URL.Query().Get("email")
	password := r.URL.Query().Get("password")
	if email == "" || password == "" {
		return apperror.BadRequestError("missing required parameters email or password")
	}

	usr, err := h.service.GetByEmailAndPassword(r.Context(), email, password)
	if err != nil {
		return err
	}

	userBytes, err := json.Marshal(usr)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(userBytes)
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) PartiallyUpdateUser(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("Partially update user")
	w.Header().Set("Content-Type", "application/json")

	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	userUUID := params.ByName("uuid")

	var updatedUser UpdateUserDTO
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			h.logger.Errorf("Error closing body %v", err)
		}
	}(r.Body)
	if err := json.NewDecoder(r.Body).Decode(&updatedUser); err != nil {
		return apperror.BadRequestError("invalid JSON scheme. check swagger API")
	}

	updatedUser.UUID = userUUID

	err := h.service.Update(r.Context(), updatedUser)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)

	return nil
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("Delete user")
	w.Header().Set("Content-Type", "application/json")

	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	userUUID := params.ByName("uuid")

	err := h.service.Delete(r.Context(), userUUID)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)

	return nil
}
