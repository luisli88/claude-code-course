//go:build integration

package persistence_test

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"myapp/internal/domain/entity"
	"myapp/internal/domain/repository"
	"myapp/internal/infrastructure/persistence"
	"myapp/internal/infrastructure/persistence/testhelper"
)

// seedTokenUser inserts a single user and returns its UUID.
func seedTokenUser(t *testing.T, db *sql.DB, email string) string {
	t.Helper()
	var id string
	err := db.QueryRow(
		"INSERT INTO users (name, email, password_hash) VALUES ($1, $2, $3) RETURNING id",
		"Token Test User", email, "$2a$10$fQhqpTNLvajhdJvuSN27m.nls7rCu4wZ.EOVxWN1n4G3st7J/wYia",
	).Scan(&id)
	if err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}
	return id
}

func cleanTokens(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec("DELETE FROM password_reset_tokens")
	if err != nil {
		t.Fatalf("failed to clean password_reset_tokens: %v", err)
	}
}

func TestPostgresPasswordResetTokenRepo_Create(t *testing.T) {
	db := testhelper.NewTestDB(t)
	cleanTokens(t, db)
	testhelper.SeedUsers(t, db, 0)
	userID := seedTokenUser(t, db, "create@test.com")
	repo := persistence.NewPostgresPasswordResetTokenRepo(db)

	token := entity.PasswordResetToken{
		UserID:    userID,
		Token:     "create-test-token",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	created, err := repo.Create(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID == 0 {
		t.Error("expected non-zero ID from DB")
	}
	if created.CreatedAt.IsZero() {
		t.Error("expected non-zero created_at from DB")
	}
	if created.Token != token.Token {
		t.Errorf("expected token %q, got %q", token.Token, created.Token)
	}
	if created.Used {
		t.Error("expected used = false on creation")
	}
}

func TestPostgresPasswordResetTokenRepo_FindByToken_NotFound(t *testing.T) {
	db := testhelper.NewTestDB(t)
	repo := persistence.NewPostgresPasswordResetTokenRepo(db)

	_, err := repo.FindByToken("nonexistent-token")
	if !errors.Is(err, repository.ErrTokenNotFound) {
		t.Errorf("expected ErrTokenNotFound, got %v", err)
	}
}

func TestPostgresPasswordResetTokenRepo_FindByToken_Found(t *testing.T) {
	db := testhelper.NewTestDB(t)
	cleanTokens(t, db)
	testhelper.SeedUsers(t, db, 0)
	userID := seedTokenUser(t, db, "find@test.com")
	repo := persistence.NewPostgresPasswordResetTokenRepo(db)

	_, err := repo.Create(entity.PasswordResetToken{
		UserID:    userID,
		Token:     "find-test-token",
		ExpiresAt: time.Now().Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	found, err := repo.FindByToken("find-test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.Token != "find-test-token" {
		t.Errorf("expected token %q, got %q", "find-test-token", found.Token)
	}
}

func TestPostgresPasswordResetTokenRepo_MarkAsUsed(t *testing.T) {
	db := testhelper.NewTestDB(t)
	cleanTokens(t, db)
	testhelper.SeedUsers(t, db, 0)
	userID := seedTokenUser(t, db, "markused@test.com")
	repo := persistence.NewPostgresPasswordResetTokenRepo(db)

	created, err := repo.Create(entity.PasswordResetToken{
		UserID:    userID,
		Token:     "markused-token",
		ExpiresAt: time.Now().Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	if err := repo.MarkAsUsed(created.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found, err := repo.FindByToken("markused-token")
	if err != nil {
		t.Fatalf("unexpected error finding token: %v", err)
	}
	if !found.Used {
		t.Error("expected used = true after MarkAsUsed")
	}
}

func TestPostgresPasswordResetTokenRepo_CountRecentByEmail(t *testing.T) {
	db := testhelper.NewTestDB(t)
	cleanTokens(t, db)
	testhelper.SeedUsers(t, db, 0)
	userID := seedTokenUser(t, db, "count@test.com")
	repo := persistence.NewPostgresPasswordResetTokenRepo(db)

	for i, tok := range []string{"count-tok-1", "count-tok-2", "count-tok-3"} {
		_, err := repo.Create(entity.PasswordResetToken{
			UserID:    userID,
			Token:     tok,
			ExpiresAt: time.Now().Add(time.Hour),
		})
		if err != nil {
			t.Fatalf("setup failed at token %d: %v", i, err)
		}
	}

	since := time.Now().Add(-time.Minute)
	count, err := repo.CountRecentByEmail("count@test.com", since)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 3 {
		t.Errorf("expected count 3, got %d", count)
	}
}

func TestPostgresPasswordResetTokenRepo_InvalidatePreviousByUserID(t *testing.T) {
	db := testhelper.NewTestDB(t)
	cleanTokens(t, db)
	testhelper.SeedUsers(t, db, 0)
	userID := seedTokenUser(t, db, "invalidate@test.com")
	repo := persistence.NewPostgresPasswordResetTokenRepo(db)

	for _, tok := range []string{"inv-tok-1", "inv-tok-2"} {
		_, err := repo.Create(entity.PasswordResetToken{
			UserID:    userID,
			Token:     tok,
			ExpiresAt: time.Now().Add(time.Hour),
		})
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
	}

	if err := repo.InvalidatePreviousByUserID(userID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, tok := range []string{"inv-tok-1", "inv-tok-2"} {
		found, err := repo.FindByToken(tok)
		if err != nil {
			t.Fatalf("unexpected error finding %q: %v", tok, err)
		}
		if !found.Used {
			t.Errorf("expected token %q to be marked as used", tok)
		}
	}
}
