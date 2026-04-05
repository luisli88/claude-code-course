package testhelper

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func NewTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping db: %v", err)
	}

	t.Cleanup(func() { db.Close() })

	return db
}

func SeedUsers(t *testing.T, db *sql.DB, count int) {
	t.Helper()

	// Clean up before seeding
	_, err := db.Exec("DELETE FROM users")
	if err != nil {
		t.Fatalf("failed to clean users table: %v", err)
	}

	for i := 1; i <= count; i++ {
		_, err := db.Exec(
			"INSERT INTO users (name, email) VALUES ($1, $2)",
			fmt.Sprintf("User %d", i),
			fmt.Sprintf("user%d@test.com", i),
		)
		if err != nil {
			t.Fatalf("failed to seed user %d: %v", i, err)
		}
	}

	t.Cleanup(func() {
		db.Exec("DELETE FROM users")
	})
}
