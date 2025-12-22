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
	userID := flag.String("id", "auth0|6943ba1bc0ceb98d69403d9c", "User ID to check")
	flag.Parse()

	fmt.Printf("🔍 Checking user: %s\n", *userID)
	fmt.Println("")

	// Check if user exists
	var userResults []json.RawMessage
	err = db.Supabase.DB.From("users").
		Select("id, email, username, created_at").
		Eq("id", *userID).
		Execute(&userResults)

	if err != nil {
		errStr := err.Error()
		if errStr == "unexpected end of JSON input" {
			fmt.Printf("❌ User %s not found in database\n", *userID)
			fmt.Println("\n💡 You need to create this user first:")
			fmt.Printf("   go run scripts/create_test_user.go -id \"%s\" -email \"user@example.com\" -username \"testuser\"\n", *userID)
		} else {
			log.Printf("❌ Error getting user: %v", err)
		}
	} else if len(userResults) == 0 {
		fmt.Printf("❌ User %s not found in database\n", *userID)
		fmt.Println("\n💡 You need to create this user first:")
		fmt.Printf("   go run scripts/create_test_user.go -id \"%s\" -email \"user@example.com\" -username \"testuser\"\n", *userID)
	} else {
		var user map[string]interface{}
		if err := json.Unmarshal(userResults[0], &user); err != nil {
			log.Printf("❌ Error unmarshalling user: %v", err)
		} else {
			fmt.Printf("✅ User found:\n")
			fmt.Printf("  ID: %s\n", user["id"])
			fmt.Printf("  Email: %s\n", user["email"])
			fmt.Printf("  Username: %s\n", user["username"])
			fmt.Println("")
		}
	}

	// List all users
	fmt.Println("📋 All users in database:")
	var allUsers []json.RawMessage
	err = db.Supabase.DB.From("users").
		Select("id, email, username").
		Execute(&allUsers)

	if err != nil {
		errStr := err.Error()
		if errStr == "unexpected end of JSON input" {
			fmt.Println("  (No users found)")
		} else {
			log.Printf("❌ Error getting all users: %v", err)
		}
	} else {
		fmt.Printf("  Total: %d user(s)\n\n", len(allUsers))
		for i, userRaw := range allUsers {
			var user map[string]interface{}
			if err := json.Unmarshal(userRaw, &user); err != nil {
				continue
			}
			fmt.Printf("  %d. %s (%s) - %s\n", i+1, user["id"], user["username"], user["email"])
		}
	}
}


