package dto

import "myapp/internal/application/usecase"

type ListUsersResponse struct {
	Users    []UserDTO `json:"users"`
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
}

type UserDTO struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func FromListUsersOutput(output *usecase.ListUsersOutput) ListUsersResponse {
	users := make([]UserDTO, len(output.Users))
	for i, u := range output.Users {
		users[i] = UserDTO{ID: u.ID, Name: u.Name, Email: u.Email}
	}
	return ListUsersResponse{
		Users:    users,
		Total:    output.Total,
		Page:     output.Page,
		PageSize: output.PageSize,
	}
}
