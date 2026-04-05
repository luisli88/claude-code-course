//go:build integration

package persistence_test

import (
	"testing"

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
