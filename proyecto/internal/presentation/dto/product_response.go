package dto

type CreateProductResponse struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Category  string  `json:"category"`
	CreatedAt string  `json:"created_at"`
}
