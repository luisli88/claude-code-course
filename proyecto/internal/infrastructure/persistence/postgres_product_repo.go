package persistence

import (
	"database/sql"
	"myapp/internal/domain/entity"
)

type PostgresProductRepo struct {
	db *sql.DB
}

func NewPostgresProductRepo(db *sql.DB) *PostgresProductRepo {
	return &PostgresProductRepo{db: db}
}

func (r *PostgresProductRepo) Create(p entity.Product) (*entity.Product, error) {
	var created entity.Product
	err := r.db.QueryRow(
		"INSERT INTO products (name, price, category) VALUES ($1, $2, $3) RETURNING id, name, price, category, created_at",
		p.Name, p.Price, p.Category,
	).Scan(&created.ID, &created.Name, &created.Price, &created.Category, &created.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &created, nil
}
