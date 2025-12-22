package main

import (
	"encoding/json"
	"fmt"
	"log"

	"meetupr-backend/internal/db"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, proceeding with environment variables")
	}

	// Initialize database
	db.Init()

	userID := "auth0|6943ba1bc0ceb98d69403d9c"
	chats, err := db.GetUserChats(userID)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Convert to JSON to see actual response
	jsonData, err := json.MarshalIndent(chats, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling JSON: %v", err)
	}

	fmt.Println("API Response JSON:")
	fmt.Println(string(jsonData))
}


