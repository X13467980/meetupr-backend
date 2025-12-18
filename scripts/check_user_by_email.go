package main

import (
	"encoding/json"
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
	email := flag.String("email", "aaaabbbb@ed.ritsumei.ac.jp", "Email address to check")
	flag.Parse()

	fmt.Printf("🔍 Checking user with email: %s\n", *email)
	fmt.Println("")

	// Check if user exists
	var userResults []json.RawMessage
	err = db.Supabase.DB.From("users").
		Select("id, email, username, created_at").
		Eq("email", *email).
		Execute(&userResults)

	if err != nil {
		errStr := err.Error()
		if errStr == "unexpected end of JSON input" {
			fmt.Printf("❌ User with email %s not found in database\n", *email)
		} else {
			log.Printf("❌ Error getting user: %v", err)
		}
		return
	}

	if len(userResults) == 0 {
		fmt.Printf("❌ User with email %s not found in database\n", *email)
		return
	}

	var user map[string]interface{}
	if err := json.Unmarshal(userResults[0], &user); err != nil {
		log.Printf("❌ Error unmarshalling user: %v", err)
		return
	}

	userID := user["id"].(string)
	fmt.Printf("✅ User found:\n")
	fmt.Printf("  ID: %s\n", userID)
	fmt.Printf("  Email: %s\n", user["email"])
	fmt.Printf("  Username: %s\n", user["username"])
	fmt.Println("")

	// Check chats for this user
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

		if len(chats) == 0 {
			fmt.Println("💡 No chats found for this user.")
			fmt.Println("   To create test chat data, run:")
			fmt.Printf("   go run scripts/create_chat_for_user.go -user \"%s\"\n", userID)
		}
	}
}
