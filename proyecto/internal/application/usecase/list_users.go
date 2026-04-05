package usecase

import (
	"myapp/internal/domain/entity"
	"myapp/internal/domain/repository"
)

type ListUsersInput struct {
	Page     int
	PageSize int
}

type ListUsersOutput struct {
	Users    []entity.User `json:"users"`
	Total    int           `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

type ListUsers struct {
	repo repository.UserRepository
}

func NewListUsers(repo repository.UserRepository) *ListUsers {
	return &ListUsers{repo: repo}
}

func (uc *ListUsers) Execute(input ListUsersInput) (*ListUsersOutput, error) {
	offset := (input.Page - 1) * input.PageSize

	users, err := uc.repo.FindAll(repository.PaginationParams{
		Limit:  input.PageSize,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	total, err := uc.repo.Count()
	if err != nil {
		return nil, err
	}

	return &ListUsersOutput{
		Users:    users,
		Total:    total,
		Page:     input.Page,
		PageSize: input.PageSize,
	}, nil
}
