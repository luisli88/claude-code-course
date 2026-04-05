package persistence

import (
	"database/sql"
	"errors"
	"myapp/internal/domain/entity"
	"myapp/internal/domain/repository"
)

type PostgresUserRepo struct {
	db *sql.DB
}

func NewPostgresUserRepo(db *sql.DB) *PostgresUserRepo {
	return &PostgresUserRepo{db: db}
}

func (r *PostgresUserRepo) FindAll(params repository.PaginationParams) ([]entity.User, error) {
	rows, err := r.db.Query(
		"SELECT id, name, email, created_at FROM users LIMIT $1 OFFSET $2",
		params.Limit, params.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var users []entity.User
	for rows.Next() {
		var u entity.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *PostgresUserRepo) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

func (r *PostgresUserRepo) FindByEmail(email string) (*entity.User, error) {
	var u entity.User
	err := r.db.QueryRow(
		"SELECT id, name, email, created_at, password_hash FROM users WHERE email = $1",
		email,
	).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.PasswordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *PostgresUserRepo) Create(user entity.User) (*entity.User, error) {
	var created entity.User
	err := r.db.QueryRow(
		"INSERT INTO users (name, email, password_hash) VALUES ($1, $2, $3) RETURNING id, name, email, created_at",
		user.Name, user.Email, user.PasswordHash,
	).Scan(&created.ID, &created.Name, &created.Email, &created.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &created, nil
}
