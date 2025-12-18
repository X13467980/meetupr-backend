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
	chatID := int64(29) // Use an existing chat ID

	fmt.Printf("🔍 Testing GetChatDetail for chat %d (user: %s)\n\n", chatID, userID)

	chat, err := db.GetChatDetail(chatID, userID)
	if err != nil {
		log.Fatalf("❌ Error: %v", err)
	}

	fmt.Printf("✅ Chat Details:\n")
	fmt.Printf("  ID: %d\n", chat.ID)
	fmt.Printf("  User1: %s\n", chat.User1ID)
	fmt.Printf("  User2: %s\n", chat.User2ID)
	if chat.AISuggestedTheme != "" {
		fmt.Printf("  Theme: %s\n", chat.AISuggestedTheme)
	}

	if chat.OtherUser != nil {
		fmt.Printf("\n👤 Other User:\n")
		fmt.Printf("  ID: %s\n", chat.OtherUser.ID)
		fmt.Printf("  Username: %s\n", chat.OtherUser.Username)
		if chat.OtherUser.Email != "" {
			fmt.Printf("  Email: %s\n", chat.OtherUser.Email)
		}
	}

	if chat.LastMessage != nil {
		fmt.Printf("\n💬 Last Message:\n")
		fmt.Printf("  ID: %d\n", chat.LastMessage.ID)
		fmt.Printf("  Content: %s\n", chat.LastMessage.Content)
		fmt.Printf("  Sender: %s\n", chat.LastMessage.SenderID)
		fmt.Printf("  Sent At: %s\n", chat.LastMessage.SentAt.Format("2006-01-02 15:04:05"))
	}

	// Convert to JSON to see actual response
	jsonData, err := json.MarshalIndent(chat, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling JSON: %v", err)
	}

	fmt.Printf("\n📄 JSON Response:\n")
	fmt.Println(string(jsonData))
}
