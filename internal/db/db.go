package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq" // PostgreSQL driver
)

var DB *sql.DB

// InitDB initializes the database connection.
func InitDB() {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	// Ping the database to verify connection
	if err = DB.Ping(); err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	log.Println("Successfully connected to the database!")
}

// CloseDB closes the database connection.
func CloseDB() {
	if DB != nil {
		DB.Close()
		log.Println("Database connection closed.")
	}
}
