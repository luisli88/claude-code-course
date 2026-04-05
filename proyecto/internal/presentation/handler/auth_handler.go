package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"myapp/internal/application/usecase"
	"myapp/internal/presentation/dto"
)

type AuthHandler struct {
	login *usecase.Login
}

func NewAuthHandler(login *usecase.Login) *AuthHandler {
	return &AuthHandler{login: login}
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
