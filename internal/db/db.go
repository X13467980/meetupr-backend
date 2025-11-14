package db

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"time"

	"meetupr-backend/internal/models"

	"github.com/nedpals/supabase-go"
)

var Supabase *supabase.Client

func Init() {
	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		log.Fatal("SUPABASE_URL environment variable not set")
	}

	supabaseKey := os.Getenv("SUPABASE_KEY")
	if supabaseKey == "" {
		log.Fatal("SUPABASE_KEY environment variable not set")
	}

	Supabase = supabase.CreateClient(supabaseURL, supabaseKey)
	log.Println("Successfully connected to Supabase")
}

func CreateUser(user models.User) error {
	var results []models.User
	err := Supabase.DB.From("users").Insert(user).Execute(&results)
	if err != nil {
		return err
	}

	// Create a corresponding profile
	profile := map[string]interface{}{
		"user_id":         user.ID,
		"native_language": "Unknown", // Default value
	}
	var profileResults []map[string]interface{}
	err = Supabase.DB.From("profiles").Insert(profile).Execute(&profileResults)
	if err != nil {
		// Rollback user creation if profile creation fails
		// This is a simplified example. In a real app, you'd want a transaction.
		Supabase.DB.From("users").Delete().Eq("id", user.ID).Execute(nil)
		return err
	}

	return nil
}

func GetUserByID(userID string) (*models.UserProfileResponse, error) {
	var results []json.RawMessage
	err := Supabase.DB.From("users").Select("*, profiles(*), user_interests(*, interests(*))").Eq("id", userID).Execute(&results)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	var user models.User
	if err := json.Unmarshal(results[0], &user); err != nil {
		return nil, err
	}

	profileResponse := &models.UserProfileResponse{
		UserID:            user.ID,
		Email:             user.Email,
		Username:          user.Username,
		Major:             user.Major,
		Gender:            user.Gender,
		NativeLanguage:    user.NativeLanguage,
		SpokenLanguages:   user.SpokenLanguages,
		LearningLanguages: user.LearningLanguages,
		Residence:         user.Residence,
		Comment:           user.Comment,
		LastUpdated:       user.LastUpdatedAt,
	}

	// Populate interests
	for _, ui := range user.Interests {
		profileResponse.Interests = append(profileResponse.Interests, models.Interest{
			ID:              ui.ID,
			Name:            ui.Name,
			PreferenceLevel: ui.PreferenceLevel,
		})
	}

	return profileResponse, nil
}

func UpdateUserProfile(userID string, req models.UpdateUserProfileRequest) (*models.UserProfileResponse, error) {
	// Update profiles table
	profileUpdate := map[string]interface{}{
		"major":              req.Major,
		"gender":             req.Gender,
		"native_language":    req.NativeLanguage,
		"spoken_languages":   req.SpokenLanguages,
		"learning_languages": req.LearningLanguages,
		"residence":          req.Residence,
		"comment":            req.Comment,
		"last_updated":       time.Now(),
	}

	var profileResults []map[string]interface{}
	err := Supabase.DB.From("profiles").Update(profileUpdate).Eq("user_id", userID).Execute(&profileResults)
	if err != nil {
		return nil, err
	}

	// Update username in users table
	userUpdate := map[string]interface{}{
		"username": req.Username,
	}
	var userResults []map[string]interface{}
	err = Supabase.DB.From("users").Update(userUpdate).Eq("id", userID).Execute(&userResults)
	if err != nil {
		return nil, err
	}

	// Update user_interests table
	// First, delete existing interests for the user
	err = Supabase.DB.From("user_interests").Delete().Eq("user_id", userID).Execute(nil)
	if err != nil {
		return nil, err
	}

	// Then, insert new interests
	var userInterests []map[string]interface{}
	for _, interestID := range req.InterestIDs {
		userInterests = append(userInterests, map[string]interface{}{
			"user_id":          userID,
			"interest_id":      interestID,
			"preference_level": 3, // Default preference level
		})
	}

	if len(userInterests) > 0 {
		err = Supabase.DB.From("user_interests").Insert(userInterests).Execute(nil)
		if err != nil {
			return nil, err
		}
	}

	// Fetch the updated profile to return
	return GetUserByID(userID)
}

func SearchUsers(interestID int, learningLanguage, spokenLanguage string) ([]models.User, error) {
	query := Supabase.DB.From("users").Select("*, profiles(*), user_interests(*, interests(*))")

	if interestID != 0 {
		query.Filter("user_interests.interest_id", "eq", strconv.Itoa(interestID))
	}

	if learningLanguage != "" {
		query.Filter("profiles.learning_languages", "cs", fmt.Sprintf("{%s}", learningLanguage))
	}

	if spokenLanguage != "" {
		query.Filter("profiles.spoken_languages", "cs", fmt.Sprintf("{%s}", spokenLanguage))
	}

	var users []models.User
	err := query.Execute(&users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func GetUserProfile(userID string) (*models.User, error) {
	var results []json.RawMessage
	err := Supabase.DB.From("users").Select("id, username, profiles(major, gender, native_language, spoken_languages, learning_languages, residence, comment), user_interests(interests(id, name), preference_level)").Eq("id", userID).Execute(&results)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	var user models.User
	if err := json.Unmarshal(results[0], &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// GetInterests returns the list of available interests (master data).
func GetInterests() ([]models.Interest, error) {
	var interests []models.Interest
	err := Supabase.DB.From("interests").Select("id, name, category").Execute(&interests)
	if err != nil {
		return nil, err
	}
	return interests, nil
}

// GetUserChats returns all chat rooms that a user is participating in
func GetUserChats(userID string) ([]models.Chat, error) {
	// Get chats where user is user1
	var results1 []json.RawMessage
	err1 := Supabase.DB.From("chats").
		Select("id, user1_id, user2_id, ai_suggested_theme, created_at").
		Eq("user1_id", userID).
		Execute(&results1)

	// Get chats where user is user2
	var results2 []json.RawMessage
	err2 := Supabase.DB.From("chats").
		Select("id, user1_id, user2_id, ai_suggested_theme, created_at").
		Eq("user2_id", userID).
		Execute(&results2)

	if err1 != nil && err2 != nil {
		return nil, fmt.Errorf("failed to get chats: %v, %v", err1, err2)
	}

	// Merge results and deduplicate by chat ID
	chatMap := make(map[int64]models.Chat)

	// Process results1
	if err1 == nil {
		for _, result := range results1 {
			var chat models.Chat
			if err := json.Unmarshal(result, &chat); err != nil {
				log.Printf("error unmarshalling chat: %v", err)
				continue
			}
			chatMap[chat.ID] = chat
		}
	}

	// Process results2
	if err2 == nil {
		for _, result := range results2 {
			var chat models.Chat
			if err := json.Unmarshal(result, &chat); err != nil {
				log.Printf("error unmarshalling chat: %v", err)
				continue
			}
			chatMap[chat.ID] = chat
		}
	}

	// Convert map to slice and populate additional info
	var chats []models.Chat
	for _, chat := range chatMap {
		// Determine the other user ID
		var otherUserID string
		if chat.User1ID == userID {
			otherUserID = chat.User2ID
		} else {
			otherUserID = chat.User1ID
		}

		// Get the other user's basic info
		otherUser, err := GetUserProfile(otherUserID)
		if err != nil {
			log.Printf("error getting other user profile: %v", err)
		} else {
			chat.OtherUser = otherUser
		}

		// Get the last message (optimized: only fetch the last one)
		lastMsg, err := GetLastChatMessage(chat.ID)
		if err == nil && lastMsg != nil {
			chat.LastMessage = lastMsg
		}

		chats = append(chats, chat)
	}

	// Sort chats by created_at descending (most recent first)
	sort.Slice(chats, func(i, j int) bool {
		return chats[i].CreatedAt.After(chats[j].CreatedAt)
	})

	return chats, nil
}

// GetLastChatMessage returns the last message in a chat room
func GetLastChatMessage(chatID int64) (*models.Message, error) {
	var messages []models.Message
	err := Supabase.DB.From("messages").
		Select("id, chat_id, sender_id, content, translated_content, message_type, sent_at").
		Eq("chat_id", strconv.FormatInt(chatID, 10)).
		Execute(&messages)
	if err != nil {
		return nil, err
	}

	if len(messages) == 0 {
		return nil, nil
	}

	// Sort messages by sent_at descending (newest first) and return the first one
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].SentAt.After(messages[j].SentAt)
	})

	return &messages[0], nil
}

// GetChatMessages returns all messages in a chat room
func GetChatMessages(chatID int64) ([]models.Message, error) {
	var messages []models.Message
	err := Supabase.DB.From("messages").
		Select("id, chat_id, sender_id, content, translated_content, message_type, sent_at").
		Eq("chat_id", strconv.FormatInt(chatID, 10)).
		Execute(&messages)
	if err != nil {
		return nil, err
	}

	// Sort messages by sent_at ascending (oldest first)
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].SentAt.Before(messages[j].SentAt)
	})

	return messages, nil
}

// IsChatParticipant checks if a user is a participant in a chat room
func IsChatParticipant(chatID int64, userID string) (bool, error) {
	// First, get the chat
	var results []json.RawMessage
	err := Supabase.DB.From("chats").
		Select("id, user1_id, user2_id").
		Eq("id", strconv.FormatInt(chatID, 10)).
		Execute(&results)
	if err != nil {
		return false, err
	}
	if len(results) == 0 {
		return false, nil
	}

	// Check if user is either user1 or user2
	var chat struct {
		ID      int64  `json:"id"`
		User1ID string `json:"user1_id"`
		User2ID string `json:"user2_id"`
	}
	if err := json.Unmarshal(results[0], &chat); err != nil {
		return false, err
	}

	return chat.User1ID == userID || chat.User2ID == userID, nil
}
