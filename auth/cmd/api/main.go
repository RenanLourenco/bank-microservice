package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	auth_proto "github.com/RenanLourenco/authentication-service/internal/auth-proto"
	"github.com/RenanLourenco/authentication-service/internal/db"
	_ "github.com/go-sql-driver/mysql"
	"google.golang.org/grpc"
)

const (
	port = "80"
	gRpcPort = "50002"
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

	app.gRPCListen()

	if err != nil {
		log.Println("Unable start the server")
		log.Panic(err)
	}

}

func buildConnectionString() string {
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbName := "auth"
	host := os.Getenv("DB_HOST")

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", user, password, host, dbName)

	fmt.Println(dsn)

	return dsn
}

func (c *Config) gRPCListen() {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%s", gRpcPort))
	if err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}

	srv := grpc.NewServer()

	auth_proto.RegisterAuthServiceServer(srv, &AuthServer{
		db: c.db,
	})

	log.Printf("gRPC server started on port %s", gRpcPort)

	if err := srv.Serve(listen); err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}
}
