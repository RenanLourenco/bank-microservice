package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/RenanLourenco/authentication-service/internal/db"
	_ "github.com/go-sql-driver/mysql"
)

const port = "80"

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

	app := Config{
		db: db,
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: app.routes(),
	}
	log.Println("Server running in port: ", port)
	err = server.ListenAndServe()

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
