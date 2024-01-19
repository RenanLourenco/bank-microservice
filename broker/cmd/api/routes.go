package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (c *Config) routes() http.Handler {
	mux := chi.NewRouter()
	// specify who is allowed to connect

	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*","http://*"},
		AllowedMethods: []string{"GET","POST","PUT","DELETE","OPTIONS"},
		AllowedHeaders: []string{"Accept","Authorization","Content-type", "X-CSRF-Token"},
		ExposedHeaders: []string{"Link"},
		AllowCredentials: true,
		MaxAge: 300,
	}))

	mux.Use(middleware.Heartbeat("/ping"))
	
	mux.Mount("/v1/transaction", c.transactionRoutes())

	mux.Route("/v1/auth", func(r chi.Router) {
		// token authentications
		r.Post("/signup", c.HandlerSignup)
		r.Post("/login", c.HandlerLogin)
		r.Post("/refresh", c.HandlerRefresh)

		// crud routes
		r.Mount("/", c.authCrudRoutes())
	})

	return mux
}

func (c *Config) transactionRoutes() http.Handler {
	mux := chi.NewRouter()
	mux.Use(middleware.DefaultLogger)
	//validation middleware below
	mux.Use(c.AuthMiddleware)
	mux.Post("/", c.HandleCreateTransaction)
	mux.Post("/deposit", c.HandleDeposit)
	return mux
}

func (c *Config) authCrudRoutes() http.Handler {
	mux := chi.NewRouter()
	mux.Use(middleware.DefaultLogger)
	mux.Use(c.AuthMiddleware)

	mux.Patch("/update/{user_id}", c.HandlerUpdateUser)

	return mux
}

// this function return all the routes that need to validade the token "Authorization:'Bearer token'"
// func (c *Config) crudRoutes() http.Handler {
// 	mux := chi.NewRouter()
// 	mux.Use(middleware.DefaultLogger)
// 	//validation middleware below
	
// 	mux.Patch("/update/{user_id}", c.UpdateUser)
// 	return mux
// }