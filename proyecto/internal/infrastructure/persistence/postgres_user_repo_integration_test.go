//go:build integration

package persistence_test

import (
	"errors"
	"testing"

	"myapp/internal/domain/entity"
	"myapp/internal/domain/repository"
	"myapp/internal/infrastructure/persistence"
	"myapp/internal/infrastructure/persistence/testhelper"
)

func TestPostgresUserRepo_FindAll(t *testing.T) {
	db := testhelper.NewTestDB(t)
	testhelper.SeedUsers(t, db, 5)
	repo := persistence.NewPostgresUserRepo(db)

	t.Run("returns all users with limit", func(t *testing.T) {
		users, err := repo.FindAll(repository.PaginationParams{Limit: 3, Offset: 0})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(users) != 3 {
			t.Errorf("expected 3 users, got %d", len(users))
		}
	})

	t.Run("returns users with offset", func(t *testing.T) {
		users, err := repo.FindAll(repository.PaginationParams{Limit: 10, Offset: 3})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(users) != 2 {
			t.Errorf("expected 2 users with offset 3, got %d", len(users))
		}
	})

	t.Run("returns empty when offset exceeds data", func(t *testing.T) {
		users, err := repo.FindAll(repository.PaginationParams{Limit: 10, Offset: 100})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(users) != 0 {
			t.Errorf("expected 0 users, got %d", len(users))
		}
	})
}

func TestPostgresUserRepo_Count(t *testing.T) {
	db := testhelper.NewTestDB(t)
	testhelper.SeedUsers(t, db, 7)
	repo := persistence.NewPostgresUserRepo(db)

	count, err := repo.Count()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 7 {
		t.Errorf("expected count 7, got %d", count)
	}
}

func TestPostgresUserRepo_Create(t *testing.T) {
	db := testhelper.NewTestDB(t)
	testhelper.SeedUsers(t, db, 0)
	repo := persistence.NewPostgresUserRepo(db)

	t.Run("creates user and returns it with DB-generated fields", func(t *testing.T) {
		user := entity.User{
			Name:         "New User",
			Email:        "new@example.com",
			PasswordHash: "$2a$10$fQhqpTNLvajhdJvuSN27m.nls7rCu4wZ.EOVxWN1n4G3st7J/wYia",
		}

		created, err := repo.Create(user)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if created.ID == "" {
			t.Error("expected non-empty ID from DB")
		}
		if created.Name != user.Name {
			t.Errorf("expected name %q, got %q", user.Name, created.Name)
		}
		if created.Email != user.Email {
			t.Errorf("expected email %q, got %q", user.Email, created.Email)
		}
		if created.CreatedAt.IsZero() {
			t.Error("expected non-zero created_at")
		}
	})

	t.Run("returns error on duplicate email", func(t *testing.T) {
		user := entity.User{
			Name:         "Duplicate",
			Email:        "new@example.com",
			PasswordHash: "$2a$10$fQhqpTNLvajhdJvuSN27m.nls7rCu4wZ.EOVxWN1n4G3st7J/wYia",
		}

		_, err := repo.Create(user)
		if err == nil {
			t.Fatal("expected error for duplicate email, got nil")
		}
	})

	t.Run("FindByEmail returns ErrNotFound for missing user", func(t *testing.T) {
		_, err := repo.FindByEmail("nobody@example.com")
		if !errors.Is(err, repository.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}
