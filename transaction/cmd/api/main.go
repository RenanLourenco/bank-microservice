package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/RenanLourenco/transaction-service/internal/db"
	transactions_proto "github.com/RenanLourenco/transaction-service/transactions-proto"
	_ "github.com/go-sql-driver/mysql"
	"google.golang.org/grpc"
)

const (
	wPort    = "80"
	gRpcPort = "50001"
)

type Config struct {
	db *db.Queries
}

func main() {
	dsn := buildConnectionString()

	dbConn, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Println("Unable to open connection with the database")
		log.Panic(err)
	}
	defer dbConn.Close()

	db := db.New(dbConn)

	app := &Config{
		db: db,
	}

	//register the grpc server
	go app.gRPCListen()

	//start web server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", wPort),
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func buildConnectionString() string {
	user := os.Getenv("TRANSACTION_DB_USER")
	password := os.Getenv("TRANSACTION_DB_PASSWORD")
	dbName := "transaction"
	host := os.Getenv("TRANSACTION_DB_HOST")

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", user, password, host, dbName)

	fmt.Println(dsn)

	return dsn
}

func (c *Config) gRPCListen() {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%s", gRpcPort))
	if err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}

	srv := grpc.NewServer()

	transactions_proto.RegisterTransactionServiceServer(srv, &TransactionServer{
		db: c.db,
	})

	log.Printf("gRPC server started on port %s", gRpcPort)

	if err := srv.Serve(listen); err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}
}
