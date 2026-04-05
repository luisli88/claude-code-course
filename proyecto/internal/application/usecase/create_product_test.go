package usecase_test

import (
	"errors"
	"testing"

	"myapp/internal/application/usecase"
	"myapp/internal/domain/repository"
)

func TestCreateProduct_Success(t *testing.T) {
	repo := repository.NewMockProductRepository()
	uc := usecase.NewCreateProduct(repo)

	product, err := uc.Execute("Laptop", 999.99, "electronics")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if product.Name != "Laptop" {
		t.Errorf("expected name Laptop, got %s", product.Name)
	}
	if product.Price != 999.99 {
		t.Errorf("expected price 999.99, got %f", product.Price)
	}
	if product.Category != "electronics" {
		t.Errorf("expected category electronics, got %s", product.Category)
	}
	if product.ID == 0 {
		t.Error("expected non-zero ID")
	}
}

func TestCreateProduct_EmptyName(t *testing.T) {
	repo := repository.NewMockProductRepository()
	uc := usecase.NewCreateProduct(repo)

	_, err := uc.Execute("", 10.0, "food")

	if !errors.Is(err, usecase.ErrInvalidProduct) {
		t.Errorf("expected ErrInvalidProduct, got %v", err)
	}
}

func TestCreateProduct_NegativePrice(t *testing.T) {
	repo := repository.NewMockProductRepository()
	uc := usecase.NewCreateProduct(repo)

	_, err := uc.Execute("Shirt", -1.0, "clothing")

	if !errors.Is(err, usecase.ErrInvalidProduct) {
		t.Errorf("expected ErrInvalidProduct, got %v", err)
	}
}

func TestCreateProduct_InvalidCategory(t *testing.T) {
	repo := repository.NewMockProductRepository()
	uc := usecase.NewCreateProduct(repo)

	_, err := uc.Execute("Gadget", 50.0, "toys")

	if !errors.Is(err, usecase.ErrInvalidProduct) {
		t.Errorf("expected ErrInvalidProduct, got %v", err)
	}
}

func TestCreateProduct_ZeroPriceAllowed(t *testing.T) {
	repo := repository.NewMockProductRepository()
	uc := usecase.NewCreateProduct(repo)

	product, err := uc.Execute("Free Sample", 0, "food")

	if err != nil {
		t.Fatalf("expected no error for zero price, got %v", err)
	}
	if product.Price != 0 {
		t.Errorf("expected price 0, got %f", product.Price)
	}
}

func TestCreateProduct_RepoError(t *testing.T) {
	repo := repository.NewMockProductRepository()
	repo.Err = errors.New("db error")
	uc := usecase.NewCreateProduct(repo)

	_, err := uc.Execute("Widget", 9.99, "electronics")

	if err == nil {
		t.Fatal("expected error from repo, got nil")
	}
}
