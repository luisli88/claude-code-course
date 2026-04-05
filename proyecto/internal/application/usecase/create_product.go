package usecase

import (
	"errors"
	"myapp/internal/domain/entity"
	"myapp/internal/domain/repository"
)

var (
	ErrInvalidProduct = errors.New("invalid product input")
)

var validCategories = map[string]bool{
	"electronics": true,
	"clothing":    true,
	"food":        true,
}

type CreateProduct struct {
	repo repository.ProductRepository
}

func NewCreateProduct(repo repository.ProductRepository) *CreateProduct {
	return &CreateProduct{repo: repo}
}

func (uc *CreateProduct) Execute(name string, price float64, category string) (*entity.Product, error) {
	if name == "" {
		return nil, ErrInvalidProduct
	}
	if price < 0 {
		return nil, ErrInvalidProduct
	}
	if !validCategories[category] {
		return nil, ErrInvalidProduct
	}

	return uc.repo.Create(entity.Product{
		Name:     name,
		Price:    price,
		Category: category,
	})
}
