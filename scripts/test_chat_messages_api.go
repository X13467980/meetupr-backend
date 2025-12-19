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

	chatID := int64(29)
	userID := "auth0|6943ba1bc0ceb98d69403d9c"

	fmt.Printf("🔍 Testing GetChatMessages for chat %d (user: %s)\n\n", chatID, userID)

	// Verify participant
	isParticipant, err := db.IsChatParticipant(chatID, userID)
	if err != nil {
		log.Fatalf("❌ Error checking participant: %v", err)
	}
	if !isParticipant {
		log.Fatalf("❌ User %s is not a participant in chat %d", userID, chatID)
	}
	fmt.Printf("✅ User is participant\n\n")

	// Get messages
	messages, err := db.GetChatMessages(chatID)
	if err != nil {
		log.Fatalf("❌ Error: %v", err)
	}

	fmt.Printf("✅ Found %d message(s):\n\n", len(messages))
	for i, msg := range messages {
		fmt.Printf("Message %d:\n", i+1)
		fmt.Printf("  ID: %d\n", msg.ID)
		fmt.Printf("  Sender: %s\n", msg.SenderID)
		fmt.Printf("  Content: %s\n", msg.Content)
		fmt.Printf("  Sent At: %s\n", msg.SentAt.Format("2006-01-02 15:04:05"))
		fmt.Println()
	}

	// Convert to JSON to see actual response
	jsonData, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling JSON: %v", err)
	}

	fmt.Printf("📄 JSON Response:\n")
	fmt.Println(string(jsonData))
}

