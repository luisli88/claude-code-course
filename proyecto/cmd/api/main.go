package main

import (
	"database/sql"
	"log"
	"myapp/internal/application/usecase"
	"myapp/internal/infrastructure/auth"
	"myapp/internal/infrastructure/persistence"
	"myapp/internal/presentation/handler"
	"myapp/internal/presentation/router"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://appuser:apppass@localhost/myapp?sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	if err := db.Ping(); err != nil {
		log.Fatalf("cannot connect to database: %v", err)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "change-me-in-production"
	}

	// Wiring: infrastructure -> application -> presentation
	userRepo := persistence.NewPostgresUserRepo(db)

	listUsers := usecase.NewListUsers(userRepo)
	userHandler := handler.NewUserHandler(listUsers)

	tokenService := auth.NewJWTTokenService(jwtSecret)
	login := usecase.NewLogin(userRepo, tokenService)
	register := usecase.NewRegister(userRepo)
	authHandler := handler.NewAuthHandler(login, register)

	r := router.New(userHandler, authHandler)

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
