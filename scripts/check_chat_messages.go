package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"meetupr-backend/internal/db"

	"github.com/joho/godotenv"
)

func containsIgnoreCase(s, substr string) bool {
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	return strings.Contains(s, substr)
}

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, proceeding with environment variables")
	}

	// Initialize database
	db.Init()

	// Check chat ID 26 (latest)
	chatID := 26
	fmt.Printf("🔍 Checking chat ID: %d\n", chatID)
	fmt.Println("")

	// Get chat info - try with json.RawMessage first
	var chatResults []json.RawMessage
	err = db.Supabase.DB.From("chats").
		Select("id, user1_id, user2_id, ai_suggested_theme, created_at").
		Eq("id", fmt.Sprintf("%d", chatID)).
		Execute(&chatResults)

	if err != nil {
		errStr := err.Error()
		// "unexpected end of JSON input"は空の結果セットを示す可能性がある
		if containsIgnoreCase(errStr, "unexpected end of json") {
			fmt.Printf("⚠️  Chat %d not found (empty result set)\n", chatID)
		} else {
			log.Printf("❌ Error getting chat: %v", err)
		}
	} else if len(chatResults) == 0 {
		fmt.Printf("⚠️  Chat %d not found (no results)\n", chatID)
	} else {
		var chat map[string]interface{}
		if len(chatResults[0]) == 0 || string(chatResults[0]) == "null" {
			fmt.Printf("⚠️  Chat %d result is empty or null\n", chatID)
		} else if err := json.Unmarshal(chatResults[0], &chat); err != nil {
			log.Printf("❌ Error unmarshalling chat: %v, raw: %s", err, string(chatResults[0]))
		} else {
			fmt.Printf("✅ Chat found:\n")
			fmt.Printf("  ID: %.0f\n", chat["id"])
			fmt.Printf("  User1: %s\n", chat["user1_id"])
			fmt.Printf("  User2: %s\n", chat["user2_id"])
			if theme, ok := chat["ai_suggested_theme"].(string); ok && theme != "" {
				fmt.Printf("  Theme: %s\n", theme)
			}
			fmt.Println("")
		}
	}

	// Get messages for this chat - try using GetChatMessages function
	fmt.Printf("📨 Checking messages for chat %d:\n", chatID)
	messages, err := db.GetChatMessages(int64(chatID))
	if err != nil {
		errStr := err.Error()
		if containsIgnoreCase(errStr, "unexpected end of json") {
			fmt.Printf("⚠️  No messages found for chat %d (empty result set)\n", chatID)
		} else {
			log.Printf("❌ Error getting messages: %v", err)
		}
	} else if len(messages) == 0 {
		fmt.Printf("⚠️  No messages found for chat %d\n", chatID)
	} else {
		fmt.Printf("✅ Found %d message(s):\n\n", len(messages))
		for i, msg := range messages {
			fmt.Printf("Message %d:\n", i+1)
			fmt.Printf("  ID: %d\n", msg.ID)
			fmt.Printf("  Sender: %s\n", msg.SenderID)
			fmt.Printf("  Content: %s\n", msg.Content)
			fmt.Printf("  Type: %s\n", msg.MessageType)
			fmt.Printf("  Sent At: %s\n", msg.SentAt.Format("2006-01-02 15:04:05"))
			fmt.Println("")
		}
	}

	// Check user's chats using GetUserChats
	userID := "auth0|6943ba1bc0ceb98d69403d9c"
	fmt.Printf("🔍 Checking chats for user: %s\n", userID)
	chats, err := db.GetUserChats(userID)
	if err != nil {
		log.Printf("❌ Error getting user chats: %v", err)
	} else {
		fmt.Printf("✅ Found %d chat(s) for user %s:\n\n", len(chats), userID)
		for i, chat := range chats {
			fmt.Printf("Chat %d:\n", i+1)
			fmt.Printf("  ID: %d\n", chat.ID)
			fmt.Printf("  User1: %s\n", chat.User1ID)
			fmt.Printf("  User2: %s\n", chat.User2ID)
			if chat.OtherUser != nil {
				fmt.Printf("  Other User: %s\n", chat.OtherUser.Username)
			}
			if chat.LastMessage != nil {
				fmt.Printf("  Last Message: %s\n", chat.LastMessage.Content)
			}
			fmt.Println("")
		}
	}
}


