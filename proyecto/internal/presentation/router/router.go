package router

import (
	"net/http"

	"myapp/internal/presentation/handler"
)

func New(userHandler *handler.UserHandler, authHandler *handler.AuthHandler, productHandler *handler.ProductHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /users", userHandler.List)
	mux.HandleFunc("POST /login", authHandler.Login)
	mux.HandleFunc("POST /register", authHandler.Register)
	mux.HandleFunc("POST /api/products", productHandler.Create)
	return mux
}
