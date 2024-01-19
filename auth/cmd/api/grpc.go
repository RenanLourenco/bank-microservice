package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"time"

	"github.com/RenanLourenco/authentication-service/external/jwt_helper"
	auth_proto "github.com/RenanLourenco/authentication-service/internal/auth-proto"
	"github.com/RenanLourenco/authentication-service/internal/db"
	transactions_proto "github.com/RenanLourenco/authentication-service/internal/transactions-proto"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthServer struct {
	auth_proto.UnimplementedAuthServiceServer
	db *db.Queries
}

func (a *AuthServer) ValidateToken(ctx context.Context, req *auth_proto.AuthValidateTokenRequest) (*auth_proto.AuthValidateResponse, error) {
	input := req.GetAuthValidateTokenEntry()

	if input.Token == "" {
		return &auth_proto.AuthValidateResponse{
			Success: false,
			Result:  "Empty token",
		}, errors.New("Received empty token")
	}

	tokenString := input.Token[len("Bearer "):]

	err := jwt_helper.VerifyToken(tokenString)
	if err != nil {
		return &auth_proto.AuthValidateResponse{
			Success: false,
			Result:  "Invalid token",
		}, errors.New("Invalid token")
	}

	return &auth_proto.AuthValidateResponse{
		Success: true,
		Result:  "Validated",
	}, nil
}

func (a *AuthServer) Signup(ctx context.Context, req *auth_proto.AuthSignupRequest) (*auth_proto.AuthSignupResponse, error) {
	input := req.GetAuthSignup()

	if input.UserType != "common" && input.UserType != "store" {
		return &auth_proto.AuthSignupResponse{
			Error:   true,
			Message: "Invalid user type, use 'common' or 'store'",
		}, errors.New("Invalid user type, use 'common' or 'store'")
	}

	findUser, _ := a.db.FindUserByEmail(ctx, input.Email)

	if findUser.ID != 0 {
		return &auth_proto.AuthSignupResponse{
			Error:   true,
			Message: "E-mail already registered",
		}, errors.New("E-mail already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
	if err != nil {
		return &auth_proto.AuthSignupResponse{
			Error:   true,
			Message: "Error generating hash",
		}, errors.New("Error generating hash")
	}

	if input.UserType == "common" {
		createUserParams := db.CreateNaturalUserParams{
			Name:     input.Name,
			Email:    input.Email,
			Password: string(hash),
			Cpf: sql.NullString{
				String: input.Cpf,
				Valid:  true,
			},
			UserType: db.UsersUserType(input.UserType),
		}

		err := a.db.CreateNaturalUser(ctx, createUserParams)

		if err != nil {
			log.Println(err)
			return &auth_proto.AuthSignupResponse{
				Error:   true,
				Message: "Error creating user",
			}, err
		}

		// jwt generation

		userConfig := jwt_helper.UsersConfig{
			Email: input.Email,
			StandardClaims: jwt.StandardClaims{
				IssuedAt:  time.Now().Unix(),
				ExpiresAt: time.Now().Add(time.Hour * 48).Unix(),
			},
		}

		signToken, err := jwt_helper.NewAccessToken(userConfig)
		refreshToken, err := jwt_helper.NewRefreshToken(userConfig.StandardClaims)

		if err != nil {
			return &auth_proto.AuthSignupResponse{
				Error:   true,
				Message: "Error generating token.",
			}, errors.New("Error generating token.")
		}

		var responseTokenData auth_proto.AuthSignupData
		responseTokenData.Token = signToken
		responseTokenData.RefreshToken = refreshToken

		err = a.createBalanceForNewUser(ctx, input.Email)
		if err != nil {
			return &auth_proto.AuthSignupResponse{
				Error:   true,
				Message: "Error creating balance.",
			},err
		}

		return &auth_proto.AuthSignupResponse{
			Error:   false,
			Message: "User created.",
			Data:    &responseTokenData,
		}, nil

	} else {
		createUserParams := db.CreateLegalUserParams{
			Name:     input.Name,
			Email:    input.Email,
			Password: string(hash),
			Cnpj: sql.NullString{
				String: input.Cnpj,
				Valid:  true,
			},
			UserType: db.UsersUserType(input.UserType),
		}

		err := a.db.CreateLegalUser(ctx, createUserParams)

		if err != nil {
			return &auth_proto.AuthSignupResponse{
				Error:   true,
				Message: "Error creating user",
			}, errors.New("Error creating user")
		}

		// jwt generation

		userConfig := jwt_helper.UsersConfig{
			Email: input.Email,
			StandardClaims: jwt.StandardClaims{
				IssuedAt:  time.Now().Unix(),
				ExpiresAt: time.Now().Add(time.Hour * 48).Unix(),
			},
		}

		signToken, err := jwt_helper.NewAccessToken(userConfig)
		refreshToken, err := jwt_helper.NewRefreshToken(userConfig.StandardClaims)

		if err != nil {
			return &auth_proto.AuthSignupResponse{
				Error:   true,
				Message: "Error generating token.",
			}, errors.New("Error generating token.")
		}

		err = a.createBalanceForNewUser(ctx, input.Email)
		if err != nil {
			return &auth_proto.AuthSignupResponse{
				Error:   true,
				Message: "Error creating balance.",
			},err
		}

		var responseTokenData auth_proto.AuthSignupData
		responseTokenData.Token = signToken
		responseTokenData.RefreshToken = refreshToken

		return &auth_proto.AuthSignupResponse{
			Error:   false,
			Message: "User created.",
			Data:    &responseTokenData,
		}, nil

	}
}

func (a *AuthServer) Login(ctx context.Context, req *auth_proto.AuthLoginRequest) (*auth_proto.AuthLoginResponse, error) {
	input := req.GetAuthLogin()

	findUser, _ := a.db.FindUserByEmail(ctx, input.Email)

	if findUser.ID == 0 {
		return &auth_proto.AuthLoginResponse{
			Error:   true,
			Message: "User not found.",
		}, errors.New("User not found.")
	}

	err := bcrypt.CompareHashAndPassword([]byte(findUser.Password), []byte(input.Password))

	if err != nil {
		return &auth_proto.AuthLoginResponse{
			Error:   true,
			Message: "Wrong password",
		}, errors.New("Wrong password")
	}

	userConfig := jwt_helper.UsersConfig{
		Email: findUser.Email,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Hour * 48).Unix(),
		},
	}

	signToken, err := jwt_helper.NewAccessToken(userConfig)
	refreshToken, err := jwt_helper.NewRefreshToken(userConfig.StandardClaims)

	if err != nil {
		log.Println(err)
		return &auth_proto.AuthLoginResponse{
			Error:   true,
			Message: "Error generating tokens",
		}, errors.New("Error generating tokens")
	}

	var responseTokenData auth_proto.AuthSignupData
	responseTokenData.Token = signToken
	responseTokenData.RefreshToken = refreshToken

	return &auth_proto.AuthLoginResponse{
		Error:   false,
		Message: "Login succeded",
		Data:    &responseTokenData,
	}, nil
}

func (a *AuthServer) Refresh(ctx context.Context, req *auth_proto.AuthRefreshRequest) (*auth_proto.AuthRefreshResponse, error) {
	input := req.GetAuthRefresh()

	userConfig := jwt_helper.ParseAccessToken(input.Token)
	refreshConfig := jwt_helper.ParseRefreshToken(input.RefreshToken)

	var refreshToken string

	if refreshConfig.Valid() != nil {
		// refresh the token
		newRefreshToken, err := jwt_helper.NewRefreshToken(*refreshConfig)
		if err != nil {
			return &auth_proto.AuthRefreshResponse{
				Error:   true,
				Message: "Error refreshing the refresh token.",
			}, errors.New("Error refreshing the refresh token.")
		}
		refreshToken = newRefreshToken
	}
	var token string
	if userConfig.StandardClaims.Valid() != nil && refreshConfig.Valid() == nil {
		newToken, err := jwt_helper.NewAccessToken(*userConfig)
		if err != nil {
			return &auth_proto.AuthRefreshResponse{
				Error:   true,
				Message: "Error refreshing the token.",
			}, errors.New("Error refreshing the token.")
		}
		token = newToken
	} else {
		token = input.Token
	}

	var responseTokenData auth_proto.AuthSignupData
	responseTokenData.Token = token
	responseTokenData.RefreshToken = refreshToken

	return &auth_proto.AuthRefreshResponse{
		Error:   false,
		Message: "Refreshed!",
		Data:    &responseTokenData,
	}, nil
}

func (a *AuthServer) UpdateUser(ctx context.Context, req *auth_proto.AuthUpdateUserRequest) (*auth_proto.AuthUpdateUserResponse, error) {
	input := req.GetAuthUpdateUser()

	findUser, _ := a.db.FindUserById(ctx, input.UserId)

	if findUser.ID == 0 {
		return &auth_proto.AuthUpdateUserResponse{
			Error:   true,
			Message: "User not found.",
		}, errors.New("User not found.")
	}

	err := a.db.UpdateUser(ctx, db.UpdateUserParams{
		Name: input.Name,
		ID:   input.UserId,
	})

	if err != nil {
		return &auth_proto.AuthUpdateUserResponse{
			Error:   true,
			Message: fmt.Sprintf("Error updating user: %s", err),
		}, errors.New(fmt.Sprintf("Error updating user: %s", err))
	}

	return &auth_proto.AuthUpdateUserResponse{
		Error:   false,
		Message: "User updated!",
	}, nil

}

func (a *AuthServer) createBalanceForNewUser(ctx context.Context, userEmail string) error {
	findUser, _ := a.db.FindUserByEmail(ctx, userEmail)

	if findUser.ID == 0 {
		return errors.New("User not found when searching to create balance.")
	}

	conn, err := grpc.Dial("transaction:50001", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	defer conn.Close()

	if err != nil {
		return errors.New("Failed to connect gRPC transaction server")
	}
	log.Println("Transaction gRPC server connected")

	client := transactions_proto.NewTransactionServiceClient(conn)

	log.Println("Sending to create balance")
	resp, err := client.CreateBalance(ctx, &transactions_proto.BalanceRequest{
		BalanceEntry: &transactions_proto.Balance{
			UserId: findUser.ID,
			Balance: 0,
		},
	})

	if resp.Success == false {
		return errors.New(resp.Result)
	}

	return nil
}
