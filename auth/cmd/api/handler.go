package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/RenanLourenco/authentication-service/external/jwt_helper"
	"github.com/RenanLourenco/authentication-service/internal/db"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

var jwt_secret = os.Getenv("jwt_secret")

func (c *Config) Signup(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var request_payload struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		CPF      string `json:"cpf,omitempty"`
		CNPJ     string `json:"cnpj,omitempty"`
		UserType string `json:"user_type"`
	}

	err := c.readJSON(w, r, &request_payload)
	if err != nil {
		c.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	if request_payload.UserType != "common" && request_payload.UserType != "store" {
		c.errorJSON(w, errors.New("Invalid user type, use 'common' or 'store'"), http.StatusBadRequest)
	}

	find_user, _ := c.db.FindUserByEmail(ctx, request_payload.Email)

	if find_user.ID != 0 {
		c.errorJSON(w, errors.New("E-mail already registered"), http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(request_payload.Password), 10)

	if err != nil {
		c.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	if request_payload.UserType == "common" {
		create_user_params := db.CreateNaturalUserParams{
			Name:     request_payload.Name,
			Email:    request_payload.Email,
			Password: string(hash),
			Balance: sql.NullString{
				Valid: true,
				String: "0",
			},
			Cpf: sql.NullString{
				String: request_payload.CPF,
				Valid:  true,
			},
			UserType: db.UsersUserType(request_payload.UserType),
		}

		err := c.db.CreateNaturalUser(ctx, create_user_params)

		if err != nil {
			c.errorJSON(w, err, http.StatusBadRequest)
			return
		}

		var response jsonResponse

		response.Error = false
		response.Message = "User created!"

		//jwt generation

		userConfig := jwt_helper.UsersConfig{
			Email: request_payload.Email,
			StandardClaims: jwt.StandardClaims{
				IssuedAt:  time.Now().Unix(),
				ExpiresAt: time.Now().Add(time.Hour * 48).Unix(),
			},
		}
		
		signToken, err := jwt_helper.NewAccessToken(userConfig)

		if err != nil {
			c.errorJSON(w, err, http.StatusBadRequest)
			return
		}
		
		response_token_map := make(map[string]string)

		response_token_map["token"] = signToken

		response.Data = response_token_map

		c.writeJSON(w, http.StatusCreated, response)
		return
	} else {
		create_user_params := db.CreateLegalUserParams{
			Name:     request_payload.Name,
			Email:    request_payload.Email,
			Password: string(hash),
			Cnpj: sql.NullString{
				String: request_payload.CNPJ,
				Valid:  true,
			},
			UserType: db.UsersUserType(request_payload.UserType),
		}

		err := c.db.CreateLegalUser(ctx, create_user_params)

		if err != nil {
			c.errorJSON(w, err, http.StatusBadRequest)
			return
		}

		var response jsonResponse

		response.Error = false
		response.Message = "User created!"

		//jwt generation

		userConfig := jwt_helper.UsersConfig{
			Email: request_payload.Email,
			StandardClaims: jwt.StandardClaims{
				IssuedAt:  time.Now().Unix(),
				ExpiresAt: time.Now().Add(time.Hour * 48).Unix(),
			},
		}
		
		signToken, err := jwt_helper.NewAccessToken(userConfig)
		refreshToken, err := jwt_helper.NewRefreshToken(userConfig.StandardClaims)

		if err != nil {
			c.errorJSON(w, err, http.StatusBadRequest)
			return
		}
		
		response_token_map := make(map[string]string)

		response_token_map["token"] = signToken
		response_token_map["refresh_token"] = refreshToken

		response.Data = response_token_map

		c.writeJSON(w, http.StatusCreated, response)
		return
	}

}

func (c *Config) Login(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var request_payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}	

	err := c.readJSON(w, r, &request_payload)
	if err != nil {
		c.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	find_user, _ := c.db.FindUserByEmail(ctx, request_payload.Email)

	if find_user.ID == 0 {
		c.errorJSON(w, errors.New("Invalid email or password"), http.StatusBadRequest)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(find_user.Password), []byte(request_payload.Password))
	if err != nil {
		c.errorJSON(w, errors.New("Wrong password!"), http.StatusBadRequest)
		return
	}

	userConfig := jwt_helper.UsersConfig{
		Email: find_user.Email,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Hour * 48).Unix(),
		},
	}

	signToken, err := jwt_helper.NewAccessToken(userConfig)
	refreshToken, err := jwt_helper.NewRefreshToken(userConfig.StandardClaims)

	if err != nil {
		c.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	var response jsonResponse

	response.Error = false
	response.Message = "Authenticated!"
	
	response_token_map := make(map[string]string)

	response_token_map["token"] = signToken
	response_token_map["refresh_token"] = refreshToken

	response.Data = response_token_map

	c.writeJSON(w, http.StatusOK, response)
	return

}

func (c *Config) Refresh(w http.ResponseWriter, r *http.Request){
	//ctx := context.Background()

	var request_payload struct {
		Token string `json:"token" binding:"required"`
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	err := c.readJSON(w, r, &request_payload)
	if err != nil {
		c.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	userConfig := jwt_helper.ParseAccessToken(request_payload.Token)
	refreshConfig := jwt_helper.ParseRefreshToken(request_payload.RefreshToken)

	if refreshConfig.Valid() != nil {
		// refresh the token
		request_payload.RefreshToken, err = jwt_helper.NewRefreshToken(*refreshConfig)
		if err != nil {
			c.errorJSON(w, errors.New("Error refreshing token"), http.StatusBadRequest)
			return
		}
	}

	if userConfig.StandardClaims.Valid() != nil && refreshConfig.Valid() == nil {
		request_payload.Token, err = jwt_helper.NewAccessToken(*userConfig)
		if err != nil {
			c.errorJSON(w, errors.New("Error refreshing access token"), http.StatusBadRequest)
			return
		}
	}

	var response jsonResponse

	response.Error = false
	response.Message = "Refreshed!"
	
	response_token_map := make(map[string]string)

	response_token_map["token"] = request_payload.Token
	response_token_map["refresh_token"] = request_payload.RefreshToken

	response.Data = response_token_map


	c.writeJSON(w,http.StatusOK,response)
	return
}

func (c *Config) UpdateUser(w http.ResponseWriter, r *http.Request) {
	// only field can be updated is the name
	ctx := context.Background()
	userIdParams := chi.URLParam(r, "user_id")

	user_id ,_ := strconv.Atoi(userIdParams)

	
	findUser, _ := c.db.FindUserById(ctx, int32(user_id))

	// if err != nil {
	// 	c.errorJSON(w, err, http.StatusBadRequest)
	// 	return
	// }

	if findUser.ID == 0 {
		c.errorJSON(w, errors.New("User not found."), http.StatusBadRequest)
		return
	}

	var request_payload struct {
		Name string `json:"name" binding:"required"`
	}

	err := c.readJSON(w,r, &request_payload)
	if err != nil {
		c.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	err = c.db.UpdateUser(ctx, db.UpdateUserParams{
		Name: request_payload.Name,
		ID: int32(user_id),
	})

	if err != nil {
		c.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	var response jsonResponse

	response.Error = false
	response.Message = "User updated!"
	c.writeJSON(w, http.StatusOK, response)
	return

}