package repository

import (
	"errors"
	"myapp/internal/domain/entity"
)

var ErrNotFound = errors.New("not found")

type PaginationParams struct {
	Limit  int
	Offset int
}

type UserRepository interface {
	FindAll(params PaginationParams) ([]entity.User, error)
	Count() (int, error)
	FindByEmail(email string) (*entity.User, error)
	Create(user entity.User) (*entity.User, error)
}
