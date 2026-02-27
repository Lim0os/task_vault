package handler

import (
	"errors"
	"net/http"
	"task_vault/internal/app/command"
	"task_vault/internal/domain"
)

type AuthHandler struct {
	register *command.RegisterUserHandler
	login    *command.LoginUserHandler
}

func NewAuthHandler(register *command.RegisterUserHandler, login *command.LoginUserHandler) *AuthHandler {
	return &AuthHandler{register: register, login: login}
}

// @Summary      Регистрация пользователя
// @Description  Создает нового пользователя по email и паролю
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input body      RegisterRequest true "Данные для регистрации"
// @Success      201   {object}  swagUserResponse
// @Failure      400   {object}  swagErrorResponse
// @Failure      409   {object}  swagErrorResponse
// @Failure      500   {object}  swagErrorResponse
// @Router       /register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "невалидный JSON")
		return
	}
	if verr := req.Validate(); verr != nil {
		writeValidationError(w, verr)
		return
	}

	user, err := h.register.Handle(r.Context(), command.RegisterUserInput{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	})
	if err != nil {
		if errors.Is(err, domain.ErrEmailTaken) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "ошибка регистрации")
		return
	}

	writeJSON(w, http.StatusCreated, UserResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
	})
}

// @Summary      Авторизация
// @Description  Возвращает JWT-токен по email и паролю
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input body      LoginRequest true "Данные для входа"
// @Success      200   {object}  swagLoginResponse
// @Failure      400   {object}  swagErrorResponse
// @Failure      401   {object}  swagErrorResponse
// @Failure      500   {object}  swagErrorResponse
// @Router       /login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "невалидный JSON")
		return
	}
	if verr := req.Validate(); verr != nil {
		writeValidationError(w, verr)
		return
	}

	output, err := h.login.Handle(r.Context(), command.LoginUserInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "ошибка авторизации")
		return
	}

	writeJSON(w, http.StatusOK, LoginResponse{Token: output.Token})
}
