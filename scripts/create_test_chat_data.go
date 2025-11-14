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
	user1ID := flag.String("user1", "", "User 1 ID (required)")
	user2ID := flag.String("user2", "", "User 2 ID (required)")
	theme := flag.String("theme", "", "AI suggested theme (optional)")
	flag.Parse()

	if *user1ID == "" || *user2ID == "" {
		log.Fatal("Both user1 and user2 IDs are required. Usage: go run scripts/create_test_chat_data.go -user1 <user1_id> -user2 <user2_id> [-theme <theme>]")
	}

	// Create chat room
	chatData := map[string]interface{}{
		"user1_id": *user1ID,
		"user2_id": *user2ID,
	}
	if *theme != "" {
		chatData["ai_suggested_theme"] = *theme
	}

	var results []map[string]interface{}
	err = db.Supabase.DB.From("chats").Insert(chatData).Execute(&results)
	if err != nil {
		log.Fatalf("Failed to create chat: %v", err)
	}

	if len(results) == 0 {
		log.Fatal("No result returned from chat creation")
	}

	chatID, ok := results[0]["id"].(float64)
	if !ok {
		log.Fatalf("Invalid chat ID in response: %v", results[0])
	}

	fmt.Printf("✅ Chat room created successfully!\n")
	fmt.Printf("   Chat ID: %.0f\n", chatID)
	fmt.Printf("   User 1: %s\n", *user1ID)
	fmt.Printf("   User 2: %s\n", *user2ID)
	if *theme != "" {
		fmt.Printf("   Theme: %s\n", *theme)
	}
	fmt.Printf("\n")
	fmt.Printf("You can now test WebSocket connection with:\n")
	fmt.Printf("  ./test_ws_client -addr localhost:8080 -chat %.0f -token \"<USER1_TOKEN>\"\n", chatID)
	fmt.Printf("  ./test_ws_client -addr localhost:8080 -chat %.0f -token \"<USER2_TOKEN>\"\n", chatID)
}

