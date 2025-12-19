package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"meetupr-backend/internal/db"

	"github.com/joho/godotenv"
)

func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, proceeding with environment variables")
	}

	// Initialize database
	db.Init()

	fmt.Println("🔍 Checking chats in database...")
	fmt.Println("")

	// Get all chats
	var allChats []json.RawMessage
	err = db.Supabase.DB.From("chats").
		Select("id, user1_id, user2_id, ai_suggested_theme, created_at").
		Execute(&allChats)

	if err != nil {
		errStr := err.Error()
		if contains(errStr, "unexpected end of json") {
			fmt.Println("⚠️  No chats found (empty result set)")
			allChats = []json.RawMessage{}
		} else {
			log.Fatalf("❌ Failed to get chats: %v", err)
		}
	}

	fmt.Printf("📊 Total chats in database: %d\n", len(allChats))
	fmt.Println("")

	for i, chatRaw := range allChats {
		var chat map[string]interface{}
		if err := json.Unmarshal(chatRaw, &chat); err != nil {
			log.Printf("⚠️  Error unmarshalling chat %d: %v", i, err)
			continue
		}

		fmt.Printf("Chat %d:\n", i+1)
		fmt.Printf("  ID: %.0f\n", chat["id"])
		fmt.Printf("  User1: %s\n", chat["user1_id"])
		fmt.Printf("  User2: %s\n", chat["user2_id"])
		if theme, ok := chat["ai_suggested_theme"].(string); ok && theme != "" {
			fmt.Printf("  Theme: %s\n", theme)
		}
		fmt.Println("")
	}

	// Check specific user's chats
	testUserID := "auth0|6917784d99703fe24aebd01d"
	fmt.Printf("🔍 Checking chats for user: %s\n", testUserID)
	fmt.Println("")

	// As user1
	var chats1 []json.RawMessage
	err1 := db.Supabase.DB.From("chats").
		Select("id, user1_id, user2_id").
		Eq("user1_id", testUserID).
		Execute(&chats1)

	if err1 != nil {
		errStr := err1.Error()
		if contains(errStr, "unexpected end of json") {
			fmt.Printf("⚠️  No chats found where user is user1 (empty result set)\n")
			chats1 = []json.RawMessage{}
		} else {
			fmt.Printf("❌ Error getting chats where user is user1: %v\n", err1)
		}
	} else {
		fmt.Printf("✅ Found %d chat(s) where user is user1\n", len(chats1))
	}

	// As user2
	var chats2 []json.RawMessage
	err2 := db.Supabase.DB.From("chats").
		Select("id, user1_id, user2_id").
		Eq("user2_id", testUserID).
		Execute(&chats2)

	if err2 != nil {
		errStr := err2.Error()
		if contains(errStr, "unexpected end of json") {
			fmt.Printf("⚠️  No chats found where user is user2 (empty result set)\n")
			chats2 = []json.RawMessage{}
		} else {
			fmt.Printf("❌ Error getting chats where user is user2: %v\n", err2)
		}
	} else {
		fmt.Printf("✅ Found %d chat(s) where user is user2\n", len(chats2))
	}

	fmt.Printf("\n📊 Total chats for user %s: %d\n", testUserID, len(chats1)+len(chats2))
}

