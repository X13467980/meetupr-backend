package main

import (
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

	fmt.Println("🔍 Checking messages in database...")
	fmt.Println("")

	// Get all messages
	var allMessages []map[string]interface{}
	err = db.Supabase.DB.From("messages").
		Select("id, chat_id, sender_id, content").
		Execute(&allMessages)

	if err != nil {
		errStr := err.Error()
		if errStr == "unexpected end of JSON input" {
			fmt.Println("⚠️  No messages found (empty result set)")
		} else {
			log.Printf("❌ Error getting messages: %v", err)
		}
		return
	}

	fmt.Printf("📨 Total messages in database: %d\n\n", len(allMessages))

	for i, msg := range allMessages {
		if i >= 10 {
			fmt.Printf("... and %d more messages\n", len(allMessages)-10)
			break
		}
		fmt.Printf("Message %d:\n", i+1)
		fmt.Printf("  ID: %.0f\n", msg["id"])
		fmt.Printf("  Chat ID: %.0f\n", msg["chat_id"])
		fmt.Printf("  Sender: %s\n", msg["sender_id"])
		fmt.Printf("  Content: %s\n", msg["content"])
		fmt.Println("")
	}

	// Check messages for specific chat IDs
	chatIDs := []int64{24, 26, 27, 28}
	for _, chatID := range chatIDs {
		fmt.Printf("📨 Messages for chat %d:\n", chatID)
		var chatMessages []map[string]interface{}
		err = db.Supabase.DB.From("messages").
			Select("id, sender_id, content").
			Eq("chat_id", fmt.Sprintf("%d", chatID)).
			Execute(&chatMessages)

		if err != nil {
			errStr := err.Error()
			if errStr == "unexpected end of JSON input" {
				fmt.Printf("  ⚠️  No messages found (empty result set)\n")
			} else {
				fmt.Printf("  ❌ Error: %v\n", err)
			}
		} else {
			fmt.Printf("  ✅ Found %d message(s)\n", len(chatMessages))
			for j, msg := range chatMessages {
				if j >= 3 {
					break
				}
				fmt.Printf("    %d. [%s] %s\n", j+1, msg["sender_id"], msg["content"])
			}
		}
		fmt.Println("")
	}
}

