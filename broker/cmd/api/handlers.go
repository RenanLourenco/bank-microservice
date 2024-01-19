package main

import (
	auth_proto "broker/internal/auth-proto"
	transactions_proto "broker/internal/transactions-proto"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type CreateTransactionPayload struct {
	Value      float64 `json:"value"`
	FromUserId int     `json:"from_user_id"`
	ToUserId   int     `json:"to_user_id"`
}

type DepositPayload struct {
	UserId int `json:"user_id"`
	Value float64 `json:"value"`
}

type SignupPayload struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	CPF      string `json:"cpf,omitempty"`
	CNPJ     string `json:"cnpj,omitempty"`
	UserType string `json:"user_type"`
}

type LoginPayload struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

type RefreshPayload struct {
	Token string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type UpdateUserPayload struct {
	Name string `json:"name"`
}

func (c *Config) HandleCreateTransaction(w http.ResponseWriter, r *http.Request) {
	var responsePayload CreateTransactionPayload
	err := c.readJSON(w, r, &responsePayload)
	if err != nil {
		c.errorJSON(w, err)
		return
	}
	//validating if user is valid to do a transaction
	_, err = validateTransaction(&responsePayload)

	if err != nil {
		c.errorJSON(w, err)
		return
	}

	err = c.pushToQueue(
		"transaction_queue",
		"transaction_topic",
		"transaction.INFO",
		responsePayload,
	)
	if err != nil {
		c.errorJSON(w, err)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Created transaction"

	c.writeJSON(w, http.StatusAccepted, payload)

}

func (c *Config) HandleDeposit(w http.ResponseWriter, r *http.Request) {
	var responsePayload DepositPayload
	err := c.readJSON(w, r, &responsePayload)
	if err != nil {
		c.errorJSON(w, err)
		return
	}
	//validating if user is valid to do a transaction
	_, err = validateDeposit(&responsePayload)

	if err != nil {
		c.errorJSON(w, err)
		return
	}

	err = c.pushToQueue(
		"deposit_queue",
		"deposit_topic",
		"deposit.INFO",
		responsePayload,
	)
	if err != nil {
		c.errorJSON(w, err)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Deposit send"

	c.writeJSON(w, http.StatusAccepted, payload)

}

func validateTransaction(payload *CreateTransactionPayload) (bool, error) {
	conn, err := grpc.Dial("transaction:50001", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	defer conn.Close()

	fmt.Println(payload)

	if err != nil {
		return false, errors.New("Failed to connect gRPC transaction server")
	}
	log.Println("Transaction gRPC server connected")

	client := transactions_proto.NewTransactionServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	log.Println("Sending to validate the transaction")
	_, err = client.VerifyIfUserValidForTransaction(ctx,&transactions_proto.VerifyIfUserValidForTransactionRequest{
		VerifyIfUserValid: &transactions_proto.VerifyIfUserValidForTransaction{
			ValueTransaction: float32(payload.Value),
			FromUserId: int32(payload.FromUserId),
			ToUserId: int32(payload.ToUserId),
		},
	})

	if err != nil {
		return false, err
	}

	return true, nil

}

func validateDeposit(payload *DepositPayload) (bool, error) {
	conn, err := grpc.Dial("transaction:50001", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	defer conn.Close()

	fmt.Println(payload)

	if err != nil {
		return false, errors.New("Failed to connect gRPC transaction server")
	}
	log.Println("Transaction gRPC server connected")

	client := transactions_proto.NewTransactionServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	log.Println("Sending to validate the deposit")

	_, err = client.VerifyIfUserValidForDeposit(ctx, &transactions_proto.VerifyIfUserValidForDepositRequest{
		VerifyIfUserValidforDeposit: &transactions_proto.VerifyIfUserValidForDeposit{
			UserId: int32(payload.UserId),
		},
	})

	if err != nil {
		return false, err
	}

	return true, nil

}


func (c *Config) HandlerSignup(w http.ResponseWriter, r *http.Request) {
	var payload SignupPayload

	err := c.readJSON(w, r, &payload)
	if err != nil {
		c.errorJSON(w, err)
		return
	}

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

	log.Println("Sending to create user")
	resp, err := client.Signup(ctx, &auth_proto.AuthSignupRequest{
		AuthSignup: &auth_proto.AuthSignup{
			Name: payload.Name,
			Email: payload.Email,
			Password: payload.Password,
			Cpf: payload.CPF,
			Cnpj: payload.CNPJ,
			UserType: payload.UserType,
		},
	})

	if err != nil {
		c.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	if resp.Error == true {
		formattedError := fmt.Sprintf("Received error from authentication: %s", resp.Message)

		c.errorJSON(w, errors.New(formattedError), http.StatusUnauthorized)
		return
	}

	c.writeJSON(w,http.StatusAccepted,resp.Data)
	return
}

func (c *Config) HandlerLogin(w http.ResponseWriter, r *http.Request){
	var payload LoginPayload

	err := c.readJSON(w, r, &payload)
	if err != nil {
		c.errorJSON(w, err)
		return
	}

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

	log.Println("Sending to do login")
	resp, err := client.Login(ctx, &auth_proto.AuthLoginRequest{
		AuthLogin: &auth_proto.AuthLogin{
			Email: payload.Email,
			Password: payload.Password,
		},
	})

	if err != nil {
		c.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	if resp.Error == true {
		formattedError := fmt.Sprintf("Received error from authentication: %s", resp.Message)

		c.errorJSON(w, errors.New(formattedError), http.StatusUnauthorized)
		return
	}


	c.writeJSON(w, http.StatusAccepted, resp.Data)
	return
}

func (c *Config) HandlerRefresh(w http.ResponseWriter, r *http.Request){
	var payload RefreshPayload
	err := c.readJSON(w, r, &payload)
	if err != nil {
		c.errorJSON(w, err)
		return
	}

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

	log.Println("Sending to refresh the token")
	resp, err := client.Refresh(ctx, &auth_proto.AuthRefreshRequest{
		AuthRefresh: &auth_proto.AuthRefresh{
			Token: payload.Token,
			RefreshToken: payload.RefreshToken,
		},
	})

	if err != nil {
		c.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	if resp.Error == true {
		formattedError := fmt.Sprintf("Received error from authentication: %s", resp.Message)

		c.errorJSON(w, errors.New(formattedError), http.StatusUnauthorized)
		return
	}

	c.writeJSON(w, http.StatusAccepted, resp.Data)
	return

}

func (c *Config) HandlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	var updateUserPayload UpdateUserPayload

	fmt.Println("chegou aqui")

	err := c.readJSON(w, r, &updateUserPayload)

	if err != nil {
		c.errorJSON(w, err)
		return
	}

	userIdParams := chi.URLParam(r, "user_id")
	userId, _ := strconv.Atoi(userIdParams)
	
	
	conn, err := grpc.Dial("auth:50002", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	defer conn.Close()

	if err != nil {
		c.errorJSON(w, errors.New(fmt.Sprintf("Failed to connect gRPC server, try again later.. %s", err)), http.StatusBadGateway)
		return
	}
	log.Println("Auth gRPC server connected")

	if err != nil {
		c.errorJSON(w, errors.New(fmt.Sprintf("Failed to connect gRPC server, try again later.. %s", err)), http.StatusBadGateway)
		return
	}
	log.Println("Auth gRPC server connected")

	client := auth_proto.NewAuthServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	log.Println("Sending to update user")
	resp, err := client.UpdateUser(ctx, &auth_proto.AuthUpdateUserRequest{
		AuthUpdateUser: &auth_proto.AuthUpdateUser{
			Name: updateUserPayload.Name,
			UserId: int32(userId),
		},
	})

	if err != nil {
		c.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	if resp.Error == true {
		formattedError := fmt.Sprintf("Received error from authentication: %s", resp.Message)

		c.errorJSON(w, errors.New(formattedError), http.StatusUnauthorized)
		return
	}



	c.writeJSON(w, http.StatusOK, resp)

}

