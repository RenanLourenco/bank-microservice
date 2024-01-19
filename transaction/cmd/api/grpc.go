package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/RenanLourenco/transaction-service/internal/db"
	transactions_proto "github.com/RenanLourenco/transaction-service/transactions-proto"
)

type TransactionServer struct {
	transactions_proto.UnimplementedTransactionServiceServer
	db *db.Queries
}

func (t *TransactionServer) CreateTransaction(ctx context.Context, req *transactions_proto.TransactionRequest) (*transactions_proto.TransactionResponse, error) {
	input := req.GetTransactionEntry()
	floatInputValue := float64(input.Value)

	fromUserBalance, err := t.db.FindBalanceByUserId(ctx,int32(input.FromUserId))
	if err != nil {
		log.Println("Failed to find balance")
		fmt.Println(err)
		return &transactions_proto.TransactionResponse{
			Success: false, Result: "Failed to find balance",
		}, err
	}
	toUserBalance, err := t.db.FindBalanceByUserId(ctx, int32(input.ToUserId))
	if err != nil {
		log.Println("Failed to find balance from to user balance")
		return &transactions_proto.TransactionResponse{
			Success: false, Result: "Failed to find balance",
		}, err
	}

	fromBalance, err := strconv.ParseFloat(fromUserBalance.Balance.String, 64)
	toBalance, err := strconv.ParseFloat(toUserBalance.Balance.String, 64)

	if fromBalance < float64(input.Value){
		log.Println("Not enought balance to do the transfer..")
		return &transactions_proto.TransactionResponse{Success: false, Result: "Not enought balance to do the transfer.."}, errors.New("Not enought balance to do the transfer..")
	}

	fromUserFinalBalance := fromBalance - floatInputValue
	toUserFinalBalance := toBalance + floatInputValue


	//update "from" balance and "to" balance

	err = t.db.UpdateBalance(ctx, db.UpdateBalanceParams{
		Balance: sql.NullString{
			String: fmt.Sprintf("%v", fromUserFinalBalance),
			Valid: true,
		},
		UserID: int32(input.FromUserId),
	})

	if err != nil{
		log.Println(err)
		return &transactions_proto.TransactionResponse{Success: false, Result: "Error updating balance"}, errors.New("Error updating balance")
	}

	err = t.db.UpdateBalance(ctx, db.UpdateBalanceParams{
		Balance: sql.NullString{
			String: fmt.Sprintf("%v", toUserFinalBalance),
			Valid: true,
		},
		UserID: int32(input.ToUserId),
	})

	if err != nil{
		return &transactions_proto.TransactionResponse{Success: false, Result: "Error updating balance"}, errors.New("Error updating balance")
	}

	// create transaction

	err = t.db.InsertTransaction(ctx, db.InsertTransactionParams{
		Value: fmt.Sprintf("%v", input.Value),
		FromUserID: fromUserBalance.UserID,
		ToUserID: toUserBalance.UserID,
	})

	if err != nil{
		return &transactions_proto.TransactionResponse{Success: false, Result: "Error inserting transaction to database"}, errors.New("Error inserting transaction to database")
	}

	return &transactions_proto.TransactionResponse{
		Success: true,
		Result: "Transaction succeded",
	}, nil

}

func (t *TransactionServer) CreateBalance(ctx context.Context, req *transactions_proto.BalanceRequest) (*transactions_proto.BalanceResponse, error) {
	input := req.GetBalanceEntry()

	err := t.db.InsertBalance(ctx, db.InsertBalanceParams{
		Balance: sql.NullString{
			Valid: true,
			String: fmt.Sprintf("%v", input.Balance),
		},
		UserID: input.UserId,
	})

	if err != nil {
		return &transactions_proto.BalanceResponse{
			Success: false,
			Result: "Error creating the balance in our database",
		},
		err
	}


	return &transactions_proto.BalanceResponse{
		Success: true,
		Result: "Balance created",
	}, nil
}

func (t *TransactionServer) Deposit(ctx context.Context, req *transactions_proto.DepositRequest) (*transactions_proto.DepositResponse, error) {
	input := req.GetDeposit()

	findUserBalance, err := t.db.FindBalanceByUserId(ctx, input.UserId)
	if err != nil {
		return &transactions_proto.DepositResponse{
			Success: false,
			Result: "User balance not found.",
		}, errors.New("User balance not found.")
	}

	userFloatBalance, err := strconv.ParseFloat(findUserBalance.Balance.String, 64) 
	if err != nil {
		return &transactions_proto.DepositResponse{
			Success: false,
			Result: "Error converting balance to float.",
		}, errors.New("Error converting balance to float.")
	}


	newBalance := userFloatBalance + float64(input.Value)

	err = t.db.UpdateBalance(ctx, db.UpdateBalanceParams{
		Balance: sql.NullString{
			String: fmt.Sprintf("%f",newBalance),
			Valid: true,
		},
		UserID: input.UserId,
	})

	if err != nil {
		return &transactions_proto.DepositResponse{
			Success: false,
			Result: "Error updating user balance.",
		}, errors.New("Error updating user balance.")
	}


	return &transactions_proto.DepositResponse{
		Success: true,
		Result: "Deposit succeded",
	}, nil

}

func (t *TransactionServer) VerifyIfUserValidForTransaction(ctx context.Context, req *transactions_proto.VerifyIfUserValidForTransactionRequest) (*transactions_proto.VerifyIfUserValidForTransactionResponse, error){
	input := req.GetVerifyIfUserValid()
	floatInputValue := float64(input.ValueTransaction)
	fmt.Println(input)

	fromUserBalance, err := t.db.FindBalanceByUserId(ctx, int32(input.FromUserId))
	if err != nil || fromUserBalance.ID == 0{
		return &transactions_proto.VerifyIfUserValidForTransactionResponse{
			Success: false,
			Result: "Failed to find balance of sender user",
		}, errors.New("Failed to find balance of sender user")
	}
	toUserBalance, err := t.db.FindBalanceByUserId(ctx, int32(input.ToUserId))
	if err != nil || toUserBalance.ID == 0{
		return &transactions_proto.VerifyIfUserValidForTransactionResponse{
			Success: false,
			Result: "Failed to find balance of receiver user",
		}, errors.New("Failed to find balance of receiver user")
	}

	fromBalance, err := strconv.ParseFloat(fromUserBalance.Balance.String, 64)

	if fromBalance < floatInputValue {
		fmt.Println("balance ok")
		return &transactions_proto.VerifyIfUserValidForTransactionResponse{
			Success: false,
			Result: "Not enought balance to do the transfer..",
		}, errors.New("Not enought balance to do the transfer..")
	}

	fmt.Println("passou de tudo")

	return &transactions_proto.VerifyIfUserValidForTransactionResponse{
		Success: true,
		Result: "Validated",
	}, nil
}

func (t *TransactionServer) VerifyIfUserValidForDeposit(ctx context.Context, req *transactions_proto.VerifyIfUserValidForDepositRequest) (*transactions_proto.VerifyIfUserValidForDepositResponse, error) {
	input := req.GetVerifyIfUserValidforDeposit()

	user, err := t.db.FindBalanceByUserId(ctx, input.UserId)

	if err != nil || user.ID == 0 {
		return &transactions_proto.VerifyIfUserValidForDepositResponse{
			Success: false,
			Result: "User balance not found.",
		}, errors.New("User balance not found.")
	}

	return &transactions_proto.VerifyIfUserValidForDepositResponse{
		Success: true,
		Result: "Valid for deposit",
	}, nil
}

