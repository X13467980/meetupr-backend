package db

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
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

// GetChats: 指定されたユーザーが参加しているチャットルーム一覧を取得します。
func GetChats(userID string) ([]models.ChatResponse, error) {
	var chats []models.ChatResponse
	// user1_id または user2_id が userID と一致するチャットを取得
	err := Supabase.DB.From("chats").
		Select("*").
		Filter("user1_id", "eq", userID).
		Execute(&chats)
	if err != nil {
		return nil, err
	}

	var chats2 []models.ChatResponse
	err = Supabase.DB.From("chats").
		Select("*").
		Filter("user2_id", "eq", userID).
		Execute(&chats2)
	if err != nil {
		return nil, err
	}

	// 両方の結果をマージ（重複排除は後処理）
	chats = append(chats, chats2...)
	return chats, nil
}

// GetChatMessages: 指定されたチャットルームのメッセージ履歴を取得します。
func GetChatMessages(chatID int64) ([]models.Message, error) {
	var messages []models.Message
	err := Supabase.DB.From("messages").
		Select("*").
		Filter("chat_id", "eq", strconv.FormatInt(chatID, 10)).
		Execute(&messages)
	if err != nil {
		return nil, err
	}
	return messages, nil
}
