package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"meetupr-backend/internal/db"
	"meetupr-backend/internal/models"

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
	userID := flag.String("id", "", "User ID (Auth0 format, e.g., auth0|xxxxx) (required)")
	email := flag.String("email", "", "Email address (required)")
	username := flag.String("username", "", "Username (required)")
	flag.Parse()

	if *userID == "" || *email == "" || *username == "" {
		log.Fatal("id, email, and username are required. Usage: go run scripts/create_test_user.go -id <user_id> -email <email> -username <username>")
	}

	// Create user
	user := models.User{
		ID:            *userID,
		Email:         *email,
		Username:      *username,
		IsOICVerified: false,
		CreatedAt:     time.Now(),
	}

	err = db.CreateUser(user)
	if err != nil {
		// Check for duplicate key errors
		if err.Error() == "PGRST202: duplicate key value violates unique constraint \"users_email_key\"" {
			log.Fatalf("❌ User with email %s already exists", *email)
		}
		if err.Error() == "PGRST202: duplicate key value violates unique constraint \"users_username_key\"" {
			log.Fatalf("❌ User with username %s already exists", *username)
		}
		if err.Error() == "PGRST202: duplicate key value violates unique constraint \"users_pkey\"" {
			log.Fatalf("❌ User with ID %s already exists", *userID)
		}
		log.Fatalf("❌ Failed to create user: %v", err)
	}

	fmt.Printf("✅ Test user created successfully!\n")
	fmt.Printf("   User ID: %s\n", *userID)
	fmt.Printf("   Email: %s\n", *email)
	fmt.Printf("   Username: %s\n", *username)
	fmt.Printf("\n")
	fmt.Printf("You can now use this user ID to create chat rooms:\n")
	fmt.Printf("  go run scripts/create_test_chat_data.go -user1 \"%s\" -user2 \"<OTHER_USER_ID>\"\n", *userID)
}

