package main

import (
	"flag"
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

	// Parse command line arguments
	chatID := flag.Int64("chat", 0, "Chat ID (required)")
	senderID := flag.String("sender", "", "Sender User ID (required)")
	content := flag.String("content", "", "Message content (required)")
	count := flag.Int("count", 1, "Number of messages to create (default: 1)")
	flag.Parse()

	if *chatID == 0 || *senderID == "" || *content == "" {
		log.Fatal("chat, sender, and content are required. Usage: go run scripts/create_test_messages.go -chat <chat_id> -sender <sender_id> -content <message> [-count <number>]")
	}

	// Create messages
	for i := 0; i < *count; i++ {
		messageData := map[string]interface{}{
			"chat_id":      *chatID,
			"sender_id":    *senderID,
			"content":      fmt.Sprintf("%s (#%d)", *content, i+1),
			"message_type": "text",
		}

		var results []map[string]interface{}
		err = db.Supabase.DB.From("messages").Insert(messageData).Execute(&results)
		if err != nil {
			log.Fatalf("Failed to create message %d: %v", i+1, err)
		}

		if len(results) > 0 {
			msgID, _ := results[0]["id"].(float64)
			fmt.Printf("✅ Message %d created (ID: %.0f)\n", i+1, msgID)
		}
	}

	fmt.Printf("\n✅ Created %d message(s) in chat %d\n", *count, *chatID)
}

