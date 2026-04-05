package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"myapp/internal/application/usecase"
	"myapp/internal/presentation/dto"
)

type ProductHandler struct {
	createProduct *usecase.CreateProduct
}

func NewProductHandler(uc *usecase.CreateProduct) *ProductHandler {
	return &ProductHandler{createProduct: uc}
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	product, err := h.createProduct.Execute(req.Name, req.Price, req.Category)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidProduct) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.CreateProductResponse{
		ID:        product.ID,
		Name:      product.Name,
		Price:     product.Price,
		Category:  product.Category,
		CreatedAt: product.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}
