package main

import (
	auth_proto "broker/internal/auth-proto"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (c *Config) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")

		conn, err := grpc.Dial("auth:50002", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
		defer conn.Close()

		if err != nil {
			c.errorJSON(w, errors.New(fmt.Sprintf("Failed to connect gRPC server, try again later.. %s", err)), http.StatusBadGateway)
			return
		}
		log.Println("Auth gRPC server connected")

		client := auth_proto.NewAuthServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		log.Println("Sending to validate the token")
		resp, err := client.ValidateToken(ctx, &auth_proto.AuthValidateTokenRequest{
			AuthValidateTokenEntry: &auth_proto.AuthValidateToken{
				Token: tokenString,
			},
		})

		if err != nil {
			c.errorJSON(w, err, http.StatusBadRequest)
			return
		}

		if resp.Success == false {
			formattedError := fmt.Sprintf("Received unauthorized from validator: %s", resp.Result)

			c.errorJSON(w, errors.New(formattedError), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)

	})
}
