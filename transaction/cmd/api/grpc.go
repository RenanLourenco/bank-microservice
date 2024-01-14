package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

	fromUserBalance, err := t.db.FindBalanceByUserId(ctx,int32(input.FromUserId))
	if err != nil {
		return &transactions_proto.TransactionResponse{
			Success: false, Result: "Failed to find balance",
		}, err
	}
	toUserBalance, err := t.db.FindBalanceByUserId(ctx, int32(input.ToUserId))
	if err != nil {
		return &transactions_proto.TransactionResponse{
			Success: false, Result: "Failed to find balance",
		}, err
	}

	fromBalance, err := strconv.ParseFloat(fromUserBalance.Balance.String, 64)
	toBalance, err := strconv.ParseFloat(toUserBalance.Balance.String, 64)

	if fromBalance < float64(input.Value){
		return &transactions_proto.TransactionResponse{Success: false, Result: "Not enought balance to do the transfer.."}, errors.New("Not enought balance to do the transfer..")
	}

	fromUserFinalBalance := fromBalance - float64(input.Value)
	toUserFinalBalance := toBalance + float64(input.Value)

	//update "from" balance and "to" balance

	err = t.db.UpdateBalance(ctx, db.UpdateBalanceParams{
		Balance: sql.NullString{
			String: fmt.Sprintf("%v", fromUserFinalBalance),
			Valid: true,
		},
		UserID: fromUserBalance.UserID,
	})

	if err != nil{
		return &transactions_proto.TransactionResponse{Success: false, Result: "Error updating balance"}, errors.New("Error updating balance")
	}

	err = t.db.UpdateBalance(ctx, db.UpdateBalanceParams{
		Balance: sql.NullString{
			String: fmt.Sprintf("%v", toUserFinalBalance),
			Valid: true,
		},
		UserID: toUserBalance.UserID,
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