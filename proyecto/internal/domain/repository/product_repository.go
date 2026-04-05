package repository

import "myapp/internal/domain/entity"

type ProductRepository interface {
	Create(p entity.Product) (*entity.Product, error)
}
