package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"myapp/internal/application/usecase"
	"myapp/internal/presentation/dto"
)

type AuthHandler struct {
	login    *usecase.Login
	register *usecase.Register
}

func NewAuthHandler(login *usecase.Login, register *usecase.Register) *AuthHandler {
	return &AuthHandler{login: login, register: register}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	output, err := h.login.Execute(usecase.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dto.LoginResponse{Token: output.Token})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	output, err := h.register.Execute(usecase.RegisterInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidInput):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, usecase.ErrEmailAlreadyTaken):
			http.Error(w, "email already taken", http.StatusConflict)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(dto.RegisterResponse{
		ID:        output.User.ID,
		Name:      output.User.Name,
		Email:     output.User.Email,
		CreatedAt: output.User.CreatedAt,
	})
}
