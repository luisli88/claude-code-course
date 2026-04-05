package usecase_test

import (
	"errors"
	"testing"
	"time"

	"myapp/internal/application/usecase"
	"myapp/internal/domain/entity"
	"myapp/internal/domain/repository"
)

func TestListUsers_Execute(t *testing.T) {
	now := time.Now()

	users := []entity.User{
		{ID: "1", Name: "Alice", Email: "alice@example.com", CreatedAt: now},
		{ID: "2", Name: "Bob", Email: "bob@example.com", CreatedAt: now},
		{ID: "3", Name: "Carol", Email: "carol@example.com", CreatedAt: now},
	}

	t.Run("returns paginated users", func(t *testing.T) {
		mock := &repository.MockUserRepository{Users: users}
		uc := usecase.NewListUsers(mock)

		output, err := uc.Execute(usecase.ListUsersInput{Page: 1, PageSize: 2})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(output.Users) != 2 {
			t.Errorf("expected 2 users, got %d", len(output.Users))
		}
		if output.Total != 3 {
			t.Errorf("expected total 3, got %d", output.Total)
		}
		if output.Page != 1 {
			t.Errorf("expected page 1, got %d", output.Page)
		}
		if output.PageSize != 2 {
			t.Errorf("expected page_size 2, got %d", output.PageSize)
		}
	})

	t.Run("returns second page", func(t *testing.T) {
		mock := &repository.MockUserRepository{Users: users}
		uc := usecase.NewListUsers(mock)

		output, err := uc.Execute(usecase.ListUsersInput{Page: 2, PageSize: 2})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(output.Users) != 1 {
			t.Errorf("expected 1 user on page 2, got %d", len(output.Users))
		}
		if output.Users[0].Name != "Carol" {
			t.Errorf("expected Carol, got %s", output.Users[0].Name)
		}
	})

	t.Run("returns empty when page exceeds data", func(t *testing.T) {
		mock := &repository.MockUserRepository{Users: users}
		uc := usecase.NewListUsers(mock)

		output, err := uc.Execute(usecase.ListUsersInput{Page: 10, PageSize: 2})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(output.Users) != 0 {
			t.Errorf("expected 0 users, got %d", len(output.Users))
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		mock := &repository.MockUserRepository{Err: errors.New("db down")}
		uc := usecase.NewListUsers(mock)

		_, err := uc.Execute(usecase.ListUsersInput{Page: 1, PageSize: 10})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "db down" {
			t.Errorf("expected 'db down', got %q", err.Error())
		}
	})
}
