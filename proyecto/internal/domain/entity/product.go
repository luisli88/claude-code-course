package entity

import "time"

type Product struct {
	ID        int
	Name      string
	Price     float64
	Category  string
	CreatedAt time.Time
}
