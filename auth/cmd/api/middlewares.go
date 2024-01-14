package main

import (
	"errors"
	"net/http"

	"github.com/RenanLourenco/authentication-service/external/jwt_helper"
)

func (c *Config) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")

		if tokenString == "" {
			c.errorJSON(w,errors.New("Need to be authenticated."),http.StatusUnauthorized)
			return
		}

		tokenString = tokenString[len("Bearer "):]

		err := jwt_helper.VerifyToken(tokenString)

		if err != nil {
			c.errorJSON(w,errors.New("Invalid token."), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w,r)
	})
}

