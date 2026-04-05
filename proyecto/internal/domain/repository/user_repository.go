package repository

import "myapp/internal/domain/entity"

type PaginationParams struct {
	Limit  int
	Offset int
}

type UserRepository interface {
	FindAll(params PaginationParams) ([]entity.User, error)
	Count() (int, error)
	FindByEmail(email string) (*entity.User, error)
}
