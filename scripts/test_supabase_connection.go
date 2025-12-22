package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"meetupr-backend/internal/db"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, proceeding with environment variables")
	}

	// Show environment variables (without exposing secrets)
	fmt.Println("🔍 Environment Variables:")
	fmt.Printf("  SUPABASE_URL: %s\n", maskURL(os.Getenv("SUPABASE_URL")))
	fmt.Printf("  SUPABASE_KEY: %s\n", maskKey(os.Getenv("SUPABASE_KEY")))
	fmt.Println("")

	// Initialize database
	db.Init()

	// Test 1: Simple query to chats table
	fmt.Println("📊 Test 1: Query all chats (no filter)")
	var allChats []json.RawMessage
	err = db.Supabase.DB.From("chats").
		Select("id, user1_id, user2_id").
		Execute(&allChats)

	if err != nil {
		fmt.Printf("  ❌ Error: %v\n", err)
	} else {
		fmt.Printf("  ✅ Success: Found %d chat(s)\n", len(allChats))
		if len(allChats) > 0 {
			var chat map[string]interface{}
			if err := json.Unmarshal(allChats[0], &chat); err == nil {
				fmt.Printf("     First chat ID: %.0f\n", chat["id"])
			}
		}
	}
	fmt.Println("")

	// Test 2: Query specific chat ID
	fmt.Println("📊 Test 2: Query chat ID 26")
	var chat26 []json.RawMessage
	err = db.Supabase.DB.From("chats").
		Select("id, user1_id, user2_id").
		Eq("id", "26").
		Execute(&chat26)

	if err != nil {
		fmt.Printf("  ❌ Error: %v\n", err)
		fmt.Printf("     Error type: %T\n", err)
	} else {
		fmt.Printf("  ✅ Success: Found %d result(s)\n", len(chat26))
		if len(chat26) > 0 {
			var chat map[string]interface{}
			if err := json.Unmarshal(chat26[0], &chat); err == nil {
				fmt.Printf("     Chat ID: %.0f\n", chat["id"])
				fmt.Printf("     User1: %s\n", chat["user1_id"])
				fmt.Printf("     User2: %s\n", chat["user2_id"])
			} else {
				fmt.Printf("     Unmarshal error: %v\n", err)
				fmt.Printf("     Raw data: %s\n", string(chat26[0]))
			}
		}
	}
	fmt.Println("")

	// Test 3: Query messages
	fmt.Println("📊 Test 3: Query messages for chat 26")
	var messages []json.RawMessage
	err = db.Supabase.DB.From("messages").
		Select("id, chat_id, sender_id, content").
		Eq("chat_id", "26").
		Execute(&messages)

	if err != nil {
		fmt.Printf("  ❌ Error: %v\n", err)
	} else {
		fmt.Printf("  ✅ Success: Found %d message(s)\n", len(messages))
		for i, msgRaw := range messages {
			if i >= 3 {
				break
			}
			var msg map[string]interface{}
			if err := json.Unmarshal(msgRaw, &msg); err == nil {
				fmt.Printf("     Message %d: ID=%.0f, Content=%s\n", i+1, msg["id"], msg["content"])
			}
		}
	}
	fmt.Println("")

	// Test 4: Count queries
	fmt.Println("📊 Test 4: Count queries")
	var countResults []map[string]interface{}
	err = db.Supabase.DB.From("chats").
		Select("id").
		Execute(&countResults)

	if err != nil {
		fmt.Printf("  ❌ Error counting chats: %v\n", err)
	} else {
		fmt.Printf("  ✅ Total chats in database: %d\n", len(countResults))
	}
}

func maskURL(url string) string {
	if url == "" {
		return "(not set)"
	}
	if len(url) > 30 {
		return url[:30] + "..."
	}
	return url
}

func maskKey(key string) string {
	if key == "" {
		return "(not set)"
	}
	if len(key) > 20 {
		return key[:10] + "..." + key[len(key)-10:]
	}
	return "***"
}


