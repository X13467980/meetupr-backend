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

	fmt.Println("🔍 Verifying chat data for user: auth0|6943ba1bc0ceb98d69403d9c")
	fmt.Println("")

	// Use GetUserChats function (which handles errors properly)
	userID := "auth0|6943ba1bc0ceb98d69403d9c"
	chats, err := db.GetUserChats(userID)
	if err != nil {
		log.Printf("❌ Error: %v", err)
	} else {
		fmt.Printf("✅ Found %d chat(s) for user %s:\n\n", len(chats), userID)
		for i, chat := range chats {
			fmt.Printf("Chat %d:\n", i+1)
			fmt.Printf("  ID: %d\n", chat.ID)
			fmt.Printf("  User1: %s\n", chat.User1ID)
			fmt.Printf("  User2: %s\n", chat.User2ID)
			if chat.AISuggestedTheme != "" {
				fmt.Printf("  Theme: %s\n", chat.AISuggestedTheme)
			}
			if chat.OtherUser != nil {
				fmt.Printf("  Other User: %s\n", chat.OtherUser.Username)
			}
			if chat.LastMessage != nil {
				fmt.Printf("  Last Message: %s\n", chat.LastMessage.Content)
			}

			// Get all messages for this chat
			messages, err := db.GetChatMessages(chat.ID)
			if err != nil {
				fmt.Printf("  ⚠️  Error getting messages: %v\n", err)
			} else {
				fmt.Printf("  Messages: %d message(s)\n", len(messages))
				for j, msg := range messages {
					if j >= 3 {
						fmt.Printf("    ... and %d more\n", len(messages)-3)
						break
					}
					fmt.Printf("    %d. [%s] %s\n", j+1, msg.SenderID, msg.Content)
				}
			}
			fmt.Println("")
		}
	}

	if len(chats) == 0 {
		fmt.Println("💡 No chats found. This could mean:")
		fmt.Println("   1. The user has no chats yet")
		fmt.Println("   2. There's an issue with the database connection")
		fmt.Println("   3. RLS policies are blocking the query")
		fmt.Println("")
		fmt.Println("💡 To create test data, run:")
		fmt.Println("   go run scripts/create_chat_for_user.go")
	}
}

