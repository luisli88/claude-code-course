package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"myapp/internal/application/usecase"
	"myapp/internal/presentation/dto"
)

type UserHandler struct {
	listUsers *usecase.ListUsers
}

func NewUserHandler(listUsers *usecase.ListUsers) *UserHandler {
	return &UserHandler{listUsers: listUsers}
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 {
		pageSize = 20
	}

	output, err := h.listUsers.Execute(usecase.ListUsersInput{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dto.FromListUsersOutput(output))
}
