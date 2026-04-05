package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"myapp/internal/application/usecase"
	"myapp/internal/domain/entity"
	"myapp/internal/domain/repository"
	"myapp/internal/presentation/dto"
	"myapp/internal/presentation/handler"
)

func TestUserHandler_List(t *testing.T) {
	now := time.Now()
	users := []entity.User{
		{ID: "1", Name: "Alice", Email: "alice@example.com", CreatedAt: now},
		{ID: "2", Name: "Bob", Email: "bob@example.com", CreatedAt: now},
	}

	t.Run("returns 200 with users", func(t *testing.T) {
		mock := &repository.MockUserRepository{Users: users}
		uc := usecase.NewListUsers(mock)
		h := handler.NewUserHandler(uc)

		req := httptest.NewRequest(http.MethodGet, "/users?page=1&page_size=10", nil)
		rec := httptest.NewRecorder()

		h.List(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		ct := rec.Header().Get("Content-Type")
		if ct != "application/json" {
			t.Errorf("expected application/json, got %q", ct)
		}

		var body dto.ListUsersResponse
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}

		if len(body.Users) != 2 {
			t.Errorf("expected 2 users, got %d", len(body.Users))
		}
		if body.Total != 2 {
			t.Errorf("expected total 2, got %d", body.Total)
		}
	})

	t.Run("defaults to page 1 and page_size 20", func(t *testing.T) {
		mock := &repository.MockUserRepository{Users: users}
		uc := usecase.NewListUsers(mock)
		h := handler.NewUserHandler(uc)

		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		rec := httptest.NewRecorder()

		h.List(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var body dto.ListUsersResponse
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}

		if body.Page != 1 {
			t.Errorf("expected default page 1, got %d", body.Page)
		}
		if body.PageSize != 20 {
			t.Errorf("expected default page_size 20, got %d", body.PageSize)
		}
	})

	t.Run("returns 500 on repository error", func(t *testing.T) {
		mock := &repository.MockUserRepository{Err: errFake}
		uc := usecase.NewListUsers(mock)
		h := handler.NewUserHandler(uc)

		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		rec := httptest.NewRecorder()

		h.List(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})
}

var errFake = &fakeError{}

type fakeError struct{}

func (e *fakeError) Error() string { return "repository failure" }
