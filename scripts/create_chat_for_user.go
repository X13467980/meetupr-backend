package main

import (
	"flag"
	"fmt"
	"log"
	"time"

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
	targetUserID := flag.String("user", "auth0|6943ba1bc0ceb98d69403d9c", "Target User ID (default: auth0|6943ba1bc0ceb98d69403d9c)")
	otherUserID := flag.String("other", "", "Other User ID (if not provided, will create a test user)")
	theme := flag.String("theme", "", "AI suggested theme (optional)")
	flag.Parse()

	// Ensure target user exists
	log.Printf("Checking if target user exists: %s", *targetUserID)
	var targetUserResults []map[string]interface{}
	err = db.Supabase.DB.From("users").
		Select("id").
		Eq("id", *targetUserID).
		Execute(&targetUserResults)

	if err != nil || len(targetUserResults) == 0 {
		log.Printf("⚠️  Target user %s not found, creating...", *targetUserID)
		userData := map[string]interface{}{
			"id":              *targetUserID,
			"email":           fmt.Sprintf("user_%s@example.com", *targetUserID),
			"username":        fmt.Sprintf("user_%s", *targetUserID),
			"is_oic_verified": false,
		}
		var createResults []map[string]interface{}
		err = db.Supabase.DB.From("users").Insert(userData).Execute(&createResults)
		if err != nil {
			errStr := err.Error()
			if errStr != "PGRST202: duplicate key value violates unique constraint \"users_pkey\"" {
				log.Fatalf("Failed to create target user: %v", err)
			}
			log.Printf("Target user already exists")
		} else {
			log.Printf("✅ Target user created: %s", *targetUserID)
		}
	} else {
		log.Printf("✅ Target user exists: %s", *targetUserID)
	}

	// If other user ID is not provided, create a test user
	if *otherUserID == "" {
		testUserID := "auth0|test_user_" + fmt.Sprintf("%d", time.Now().Unix())
		testEmail := fmt.Sprintf("test_%d@example.com", time.Now().Unix())
		testUsername := fmt.Sprintf("testuser_%d", time.Now().Unix())

		log.Printf("Creating test user: %s", testUserID)
		userData := map[string]interface{}{
			"id":              testUserID,
			"email":           testEmail,
			"username":        testUsername,
			"is_oic_verified": false,
		}

		var userResults []map[string]interface{}
		err = db.Supabase.DB.From("users").Insert(userData).Execute(&userResults)
		if err != nil {
			// Check if user already exists
			errStr := err.Error()
			if errStr != "PGRST202: duplicate key value violates unique constraint \"users_pkey\"" {
				log.Fatalf("Failed to create test user: %v", err)
			}
			log.Printf("Test user already exists, using existing user")
		} else {
			log.Printf("✅ Test user created: %s", testUserID)
		}

		*otherUserID = testUserID
	} else {
		// Ensure other user exists
		log.Printf("Checking if other user exists: %s", *otherUserID)
		var otherUserResults []map[string]interface{}
		err = db.Supabase.DB.From("users").
			Select("id").
			Eq("id", *otherUserID).
			Execute(&otherUserResults)

		if err != nil || len(otherUserResults) == 0 {
			log.Fatalf("Other user %s not found. Please create it first or use -other without value to auto-create", *otherUserID)
		} else {
			log.Printf("✅ Other user exists: %s", *otherUserID)
		}
	}

	// Check if chat already exists
	log.Printf("Checking if chat already exists between %s and %s", *targetUserID, *otherUserID)
	var existingChats1 []map[string]interface{}
	var existingChats2 []map[string]interface{}

	err1 := db.Supabase.DB.From("chats").
		Select("id").
		Eq("user1_id", *targetUserID).
		Eq("user2_id", *otherUserID).
		Execute(&existingChats1)

	err2 := db.Supabase.DB.From("chats").
		Select("id").
		Eq("user1_id", *otherUserID).
		Eq("user2_id", *targetUserID).
		Execute(&existingChats2)

	var chatID int64
	if err1 == nil && len(existingChats1) > 0 {
		chatID = int64(existingChats1[0]["id"].(float64))
		log.Printf("✅ Found existing chat ID: %d", chatID)
	} else if err2 == nil && len(existingChats2) > 0 {
		chatID = int64(existingChats2[0]["id"].(float64))
		log.Printf("✅ Found existing chat ID: %d", chatID)
	} else {
		// Create chat room
		log.Printf("Creating new chat room between %s and %s", *targetUserID, *otherUserID)
		chatData := map[string]interface{}{
			"user1_id": *targetUserID,
			"user2_id": *otherUserID,
		}
		if *theme != "" {
			chatData["ai_suggested_theme"] = *theme
		}

		var chatResults []map[string]interface{}
		err = db.Supabase.DB.From("chats").Insert(chatData).Execute(&chatResults)
		if err != nil {
			log.Fatalf("Failed to create chat: %v", err)
		}

		if len(chatResults) == 0 {
			log.Fatal("No result returned from chat creation")
		}

		chatID = int64(chatResults[0]["id"].(float64))
		log.Printf("✅ Chat room created successfully! Chat ID: %d", chatID)
	}

	fmt.Printf("\n✅ Chat room ready!\n")
	fmt.Printf("   Chat ID: %d\n", chatID)
	fmt.Printf("   User 1: %s\n", *targetUserID)
	fmt.Printf("   User 2: %s\n", *otherUserID)
	if *theme != "" {
		fmt.Printf("   Theme: %s\n", *theme)
	}

	// Create sample messages
	createSampleMessages(chatID, *targetUserID, *otherUserID)

	fmt.Printf("\n✅ Chat data creation completed!\n")
	fmt.Printf("\nYou can test the chat with:\n")
	fmt.Printf("  GET http://localhost:8080/api/v1/chats (with JWT token for user %s)\n", *targetUserID)
	fmt.Printf("  GET http://localhost:8080/api/v1/chats/%d/messages (with JWT token)\n", chatID)
	fmt.Printf("  WS ws://localhost:8080/ws/chat/%d (with JWT token)\n", chatID)
}

func createSampleMessages(chatID int64, user1ID, user2ID string) {
	log.Printf("Creating sample messages for chat %d", chatID)

	messages := []struct {
		senderID string
		content  string
	}{
		{user1ID, "こんにちは！"},
		{user2ID, "Hello! Nice to meet you."},
		{user1ID, "よろしくお願いします！"},
		{user2ID, "Let's chat!"},
		{user1ID, "今日はいい天気ですね"},
		{user2ID, "Yes, it's a beautiful day!"},
	}

	for i, msg := range messages {
		messageData := map[string]interface{}{
			"chat_id":      chatID,
			"sender_id":    msg.senderID,
			"content":      msg.content,
			"message_type": "text",
		}

		var results []map[string]interface{}
		err := db.Supabase.DB.From("messages").Insert(messageData).Execute(&results)
		if err != nil {
			log.Printf("⚠️  Failed to create message %d: %v", i+1, err)
			continue
		}

		if len(results) > 0 {
			msgID, _ := results[0]["id"].(float64)
			log.Printf("✅ Message %d created (ID: %.0f): %s", i+1, msgID, msg.content)
		}
		// Add small delay to ensure different timestamps
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("✅ Created %d sample messages\n", len(messages))
}

