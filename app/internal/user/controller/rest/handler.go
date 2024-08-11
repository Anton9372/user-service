package rest

import (
	"Users/internal/apperror"
	h "Users/internal/handler"
	"Users/internal/user/controller"
	"Users/internal/user/domain/dto"
	"Users/pkg/logging"
	"Users/pkg/utils"
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

type handler struct {
	service controller.Service
	logger  *logging.Logger
}

func NewHandler(service controller.Service, logger *logging.Logger) h.Handler {
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

// CreateUser
// @Summary 	Create user
// @Description Creates new user
// @Tags 		User
// @Accept		json
// @Param 		input	body 	 user.CreateUserDTO	true	"User's data"
// @Success 	201
// @Failure 	400 	{object} apperror.AppError "Validation error"
// @Failure 	418 	{object} apperror.AppError "Something wrong with application logic"
// @Failure 	500 	{object} apperror.AppError "Internal server error"
// @Router /users [post]
func (h *handler) CreateUser(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("Create user")
	defer utils.CloseBody(h.logger, r.Body)
	w.Header().Set("Content-Type", "application/json")

	var createdUser dto.CreateUserDTO

	if err := json.NewDecoder(r.Body).Decode(&createdUser); err != nil {
		return apperror.BadRequestError("invalid JSON scheme. check swagger API")
	}

	errEmpty := createdUser.ValidateEmptyFields()
	if errEmpty != nil {
		return apperror.BadRequestError(errEmpty.Error())
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

// GetAllUsers
// @Summary 	Get all users
// @Description Get list of all users
// @Tags 		User
// @Produce 	json
// @Success 	200		{object} []user.User "Users list"
// @Failure 	418 	{object} apperror.AppError "Something wrong with application logic"
// @Failure 	500 	{object} apperror.AppError "Internal server error"
// @Router 		/users/all 		[get]
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

// GetUserByUUID
// @Summary 	Get user by uuid
// @Description Get user by uuid
// @Tags 		User
// @Produce 	json
// @Param 		uuid 	path 	 string 	true  "User's uuid"
// @Success 	200		{object} user.User "User"
// @Failure 	404 	{object} apperror.AppError "User not found"
// @Failure 	418 	{object} apperror.AppError "Something wrong with application logic"
// @Failure 	500 	{object} apperror.AppError "Internal server error"
// @Router 		/users/one	[get]
func (h *handler) GetUserByUUID(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("Get user by uuid")
	defer utils.CloseBody(h.logger, r.Body)
	w.Header().Set("Content-Type", "application/json")

	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	userUUID := params.ByName("uuid")
	if userUUID == "" {
		return apperror.BadRequestError("user uuid must not be empty")
	}

	user, err := h.service.GetByUUID(r.Context(), userUUID)
	if err != nil {
		return err
	}

	userBytes, err := json.Marshal(user)
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

// GetUserByEmailAndPassword
// @Summary 	Get user by email and password
// @Description Get user by email and password
// @Tags 		User
// @Produce 	json
// @Param 		email 		path 	 string 	true  "User's email"
// @Param 		password 	path 	 string 	true  "User's password"
// @Success 	200		{object} user.User "User"
// @Failure 	404 	{object} apperror.AppError "User not found"
// @Failure 	418 	{object} apperror.AppError "Something wrong with application logic"
// @Failure 	500 	{object} apperror.AppError "Internal server error"
// @Router 		/users	[get]
func (h *handler) GetUserByEmailAndPassword(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("Get user by email and password")
	defer utils.CloseBody(h.logger, r.Body)
	w.Header().Set("Content-Type", "application/json")

	email := r.URL.Query().Get("email")
	password := r.URL.Query().Get("password")
	if email == "" {
		return apperror.BadRequestError("email must not be empty")
	}
	if password == "" {
		return apperror.BadRequestError("password must not be empty")
	}

	user, err := h.service.GetByEmailAndPassword(r.Context(), email, password)
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

// PartiallyUpdateUser
// @Summary 	Update user
// @Description Update user
// @Tags 		User
// @Accept		json
// @Param 		user_uuid 	path 	 string 			true  "User's uuid"
// @Param 		input 		body 	 user.UpdateUserDTO true  "User's data"
// @Success 	204
// @Failure 	400 	{object} apperror.AppError "Validation error"
// @Failure 	418 	{object} apperror.AppError "Something wrong with application logic"
// @Failure 	500 	{object} apperror.AppError "Internal server error"
// @Router /users/one [patch]
func (h *handler) PartiallyUpdateUser(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("Partially update user")
	defer utils.CloseBody(h.logger, r.Body)
	w.Header().Set("Content-Type", "application/json")

	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	userUUID := params.ByName("uuid")

	var updatedUser dto.UpdateUserDTO
	if err := json.NewDecoder(r.Body).Decode(&updatedUser); err != nil {
		return apperror.BadRequestError("invalid JSON scheme. check swagger API")
	}

	updatedUser.UUID = userUUID

	if err := updatedUser.ValidateEmptyFields(); err != nil {
		return apperror.BadRequestError(err.Error())
	}

	err := h.service.Update(r.Context(), updatedUser)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)

	h.logger.Info("Partially update user successfully")
	return nil
}

// DeleteUser
// @Summary 	Delete user
// @Description Delete user
// @Tags 		User
// @Param 		user_uuid 	path 	 string 			true  "User's uuid"
// @Success 	204
// @Failure 	404 	{object} apperror.AppError "user not found"
// @Failure 	418 	{object} apperror.AppError "Something wrong with application logic"
// @Failure 	500 	{object} apperror.AppError "Internal server error"
// @Router /users/one [delete]
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
