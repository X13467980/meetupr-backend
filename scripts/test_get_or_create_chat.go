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

	user1ID := "auth0|6943ba1bc0ceb98d69403d9c"
	user2ID := "auth0|test_user_1766072007"

	fmt.Printf("🔍 Testing GetOrCreateChat between %s and %s\n\n", user1ID, user2ID)

	chatID, err := db.GetOrCreateChat(user1ID, user2ID)
	if err != nil {
		log.Fatalf("❌ Error: %v", err)
	}

	fmt.Printf("✅ Chat ID: %d\n\n", chatID)

	// Get chat details
	chat, err := db.GetChatDetail(chatID, user1ID)
	if err != nil {
		log.Fatalf("❌ Error getting chat details: %v", err)
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(chat, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling JSON: %v", err)
	}

	fmt.Printf("📄 Chat Details:\n")
	fmt.Println(string(jsonData))
}
