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
	userID := flag.String("id", "auth0|6926bbc901ffb16fd8a83f6b", "User ID to check")
	flag.Parse()

	fmt.Printf("🔍 Checking avatar_url for user: %s\n\n", *userID)

	// Get avatar_url from profiles table (individual field query to avoid JSON parsing issues)
	var avatarResults []map[string]interface{}
	err = db.Supabase.DB.From("profiles").
		Select("avatar_url").
		Eq("user_id", *userID).
		Execute(&avatarResults)

	if err != nil {
		errStr := err.Error()
		if errStr == "unexpected end of JSON input" {
			fmt.Printf("❌ No profile found for user %s (or profile exists but avatar_url query failed)\n", *userID)

			// Try to check if profile exists at all
			var profileExists []map[string]interface{}
			err2 := db.Supabase.DB.From("profiles").
				Select("user_id").
				Eq("user_id", *userID).
				Execute(&profileExists)

			if err2 == nil && len(profileExists) > 0 {
				fmt.Printf("⚠️  Profile exists but avatar_url query failed (might be NULL)\n")
			} else {
				fmt.Printf("❌ Profile does not exist for this user\n")
			}
		} else {
			log.Printf("❌ Error querying profiles: %v", err)
		}
		return
	}

	if len(avatarResults) == 0 {
		fmt.Printf("❌ No profile found for user %s\n", *userID)
		return
	}

	profile := avatarResults[0]
	avatarURL, ok := profile["avatar_url"]

	fmt.Printf("Profile found:\n")
	fmt.Printf("  user_id: %s\n", *userID)
	fmt.Printf("  avatar_url: %v\n", avatarURL)

	if !ok || avatarURL == nil {
		fmt.Printf("\n❌ avatar_url is NULL\n")
		fmt.Printf("\n💡 This means the search API will return: \"avatar_url\": null\n")
	} else if avatarURLStr, ok := avatarURL.(string); ok {
		if avatarURLStr == "" {
			fmt.Printf("\n❌ avatar_url is empty string\n")
		} else {
			fmt.Printf("\n✅ avatar_url found: %s\n", avatarURLStr)

			// Pretty print JSON for testing (as it would appear in search API response)
			testResult := map[string]interface{}{
				"user_id":    *userID,
				"avatar_url": avatarURLStr,
			}
			jsonBytes, _ := json.MarshalIndent(testResult, "", "  ")
			fmt.Printf("\n📤 As it would appear in search API response:\n%s\n", string(jsonBytes))
		}
	} else {
		fmt.Printf("\n⚠️  avatar_url is not a string: %T (%v)\n", avatarURL, avatarURL)
	}
}

