package repository

import (
	"myapp/internal/domain/entity"
	"time"
)

type MockProductRepository struct {
	Products []entity.Product
	Err      error
}

func NewMockProductRepository() *MockProductRepository {
	return &MockProductRepository{}
}

func (m *MockProductRepository) Create(p entity.Product) (*entity.Product, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	p.ID = len(m.Products) + 1
	p.CreatedAt = time.Now()
	m.Products = append(m.Products, p)
	return &p, nil
}
