package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"meetupr-backend/internal/db"
	"meetupr-backend/internal/models"

	"github.com/joho/godotenv"
)

// contains checks if a string contains a substring (case-insensitive)
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

	fmt.Println("🌱 Starting database seeding...")
	fmt.Println("")

	// 1. Create test users
	fmt.Println("👥 Creating test users...")
	users := []struct {
		ID       string
		Email    string
		Username string
	}{
		{"auth0|6917784d99703fe24aebd01d", "testuser1@example.com", "testuser1"},
		{"auth0|test_user_67890", "testuser2@example.com", "testuser2"},
		{"auth0|test_user_11111", "testuser3@example.com", "testuser3"},
		{"auth0|test_user_22222", "testuser4@example.com", "testuser4"},
	}

	createdUsers := []string{}
	for _, u := range users {
		user := models.User{
			ID:            u.ID,
			Email:         u.Email,
			Username:      u.Username,
			IsOICVerified: false,
			CreatedAt:     time.Now(),
		}

		err := db.CreateUser(user)
		if err != nil {
			errStr := err.Error()
			// Check if user already exists (various error formats)
			if contains(errStr, "duplicate key") || contains(errStr, "23505") || contains(errStr, "23503") {
				fmt.Printf("  ⚠️  User %s already exists, skipping...\n", u.Username)
				createdUsers = append(createdUsers, u.ID)
				continue
			}
			log.Fatalf("❌ Failed to create user %s: %v", u.Username, err)
		}
		fmt.Printf("  ✅ Created user: %s (%s)\n", u.Username, u.ID)
		createdUsers = append(createdUsers, u.ID)
	}
	fmt.Println("")

	// 2. Create chat rooms
	fmt.Println("💬 Creating chat rooms...")
	chats := []struct {
		User1ID string
		User2ID string
		Theme   string
		ChatID  int64 // Will be set after creation
	}{
		{createdUsers[0], createdUsers[1], "ゲームについて話そう", 0},
		{createdUsers[0], createdUsers[2], "プログラミングの勉強", 0},
		{createdUsers[1], createdUsers[2], "旅行の計画", 0},
		{createdUsers[2], createdUsers[3], "映画の感想", 0},
	}

	for i := range chats {
		c := &chats[i]
		chatData := map[string]interface{}{
			"user1_id":           c.User1ID,
			"user2_id":           c.User2ID,
			"ai_suggested_theme": c.Theme,
		}

		var results []map[string]interface{}
		err := db.Supabase.DB.From("chats").Insert(chatData).Execute(&results)
		if err != nil {
			errStr := err.Error()
			// Check if chat already exists (unique constraint)
			if contains(errStr, "duplicate key") || contains(errStr, "23505") || contains(errStr, "idx_unique_chat_pair") {
				fmt.Printf("  ⚠️  Chat between %s and %s already exists, trying to find...\n", c.User1ID, c.User2ID)
				// Try to get existing chat ID directly from Supabase
				var existingChats []map[string]interface{}
				// Try user1_id = c.User1ID AND user2_id = c.User2ID
				err1 := db.Supabase.DB.From("chats").
					Select("id").
					Eq("user1_id", c.User1ID).
					Eq("user2_id", c.User2ID).
					Execute(&existingChats)

				if err1 == nil && len(existingChats) > 0 {
					if id, ok := existingChats[0]["id"].(float64); ok {
						c.ChatID = int64(id)
						fmt.Printf("  ✅ Found existing chat room (ID: %d)\n", c.ChatID)
						continue
					}
				}

				// Try user1_id = c.User2ID AND user2_id = c.User1ID (reverse order)
				var existingChats2 []map[string]interface{}
				err2 := db.Supabase.DB.From("chats").
					Select("id").
					Eq("user1_id", c.User2ID).
					Eq("user2_id", c.User1ID).
					Execute(&existingChats2)

				if err2 == nil && len(existingChats2) > 0 {
					if id, ok := existingChats2[0]["id"].(float64); ok {
						c.ChatID = int64(id)
						fmt.Printf("  ✅ Found existing chat room (ID: %d)\n", c.ChatID)
						continue
					}
				}

				fmt.Printf("  ⚠️  Could not find existing chat ID, skipping...\n")
				continue
			}
			log.Fatalf("❌ Failed to create chat: %v", err)
		}

		if len(results) > 0 {
			chatID, ok := results[0]["id"].(float64)
			if !ok {
				log.Fatalf("❌ Invalid chat ID in response: %v", results[0])
			}
			c.ChatID = int64(chatID)
			fmt.Printf("  ✅ Created chat room (ID: %.0f) between %s and %s\n", chatID, c.User1ID, c.User2ID)
		}
	}
	fmt.Println("")

	// 3. Create messages
	fmt.Println("📝 Creating messages...")
	messages := []struct {
		ChatIndex int // Index in chats array
		SenderID  string
		Content   string
	}{
		// Chat 1 messages (between user1 and user2)
		{0, createdUsers[0], "こんにちは！"},
		{0, createdUsers[1], "こんにちは！よろしくお願いします。"},
		{0, createdUsers[0], "ゲームは何が好きですか？"},
		{0, createdUsers[1], "RPGが好きです！あなたは？"},
		{0, createdUsers[0], "アクションゲームが好きです！"},

		// Chat 2 messages (between user1 and user3)
		{1, createdUsers[0], "プログラミングの勉強を始めました"},
		{1, createdUsers[2], "いいですね！何の言語を勉強していますか？"},
		{1, createdUsers[0], "Go言語です！"},

		// Chat 3 messages (between user2 and user3)
		{2, createdUsers[1], "旅行に行きたいです"},
		{2, createdUsers[2], "どこに行きたいですか？"},

		// Chat 4 messages (between user3 and user4)
		{3, createdUsers[2], "最近見た映画が面白かったです"},
		{3, createdUsers[3], "どんな映画ですか？"},
	}

	messageCount := 0
	for _, m := range messages {
		// Skip if chat index is out of range or chat doesn't exist
		if m.ChatIndex >= len(chats) || chats[m.ChatIndex].ChatID == 0 {
			continue
		}

		chatID := chats[m.ChatIndex].ChatID
		messageData := map[string]interface{}{
			"chat_id":      chatID,
			"sender_id":    m.SenderID,
			"content":      m.Content,
			"message_type": "text",
		}

		var results []map[string]interface{}
		err := db.Supabase.DB.From("messages").Insert(messageData).Execute(&results)
		if err != nil {
			log.Printf("⚠️  Failed to create message: %v", err)
			continue
		}

		if len(results) > 0 {
			messageCount++
		}
	}
	fmt.Printf("  ✅ Created %d message(s)\n", messageCount)
	fmt.Println("")

	// Count created chats
	createdChatCount := 0
	for _, c := range chats {
		if c.ChatID > 0 {
			createdChatCount++
		}
	}

	// Summary
	fmt.Println("✨ Seeding completed!")
	fmt.Println("")
	fmt.Println("📊 Summary:")
	fmt.Printf("  - Users: %d\n", len(createdUsers))
	fmt.Printf("  - Chat rooms: %d\n", createdChatCount)
	fmt.Printf("  - Messages: %d\n", messageCount)
	fmt.Println("")
	fmt.Println("🧪 Test the API with:")
	fmt.Printf("  ./scripts/test_api.sh \"%s\" \"testuser1@example.com\"\n", createdUsers[0])
	fmt.Printf("  ./scripts/test_api.sh \"%s\" \"testuser2@example.com\"\n", createdUsers[1])
}
