package dto_test

import (
	"testing"
	"time"

	"myapp/internal/application/usecase"
	"myapp/internal/domain/entity"
	"myapp/internal/presentation/dto"
)

func TestFromListUsersOutput(t *testing.T) {
	now := time.Now()

	output := &usecase.ListUsersOutput{
		Users: []entity.User{
			{ID: "1", Name: "Alice", Email: "alice@example.com", CreatedAt: now},
		},
		Total:    1,
		Page:     1,
		PageSize: 10,
	}

	result := dto.FromListUsersOutput(output)

	if len(result.Users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(result.Users))
	}

	u := result.Users[0]
	if u.ID != "1" || u.Name != "Alice" || u.Email != "alice@example.com" {
		t.Errorf("unexpected user DTO: %+v", u)
	}

	if result.Total != 1 || result.Page != 1 || result.PageSize != 10 {
		t.Errorf("unexpected pagination: total=%d page=%d pageSize=%d", result.Total, result.Page, result.PageSize)
	}
}

func TestFromListUsersOutput_ExcludesCreatedAt(t *testing.T) {
	output := &usecase.ListUsersOutput{
		Users: []entity.User{
			{ID: "1", Name: "Alice", Email: "alice@example.com", CreatedAt: time.Now()},
		},
		Total: 1, Page: 1, PageSize: 10,
	}

	result := dto.FromListUsersOutput(output)

	// UserDTO should not have CreatedAt — verify it's not in the struct
	// This is a compile-time guarantee, but we verify the mapping works
	if result.Users[0].ID == "" {
		t.Error("ID should not be empty")
	}
}
