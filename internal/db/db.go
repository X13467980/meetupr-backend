package db

import (
	"log"
	"os"
	"time"

	"github.com/nedpals/supabase-go"
	"meetupr-backend/internal/models"
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
	var user models.User
	err := Supabase.DB.From("users").Select("*, profiles(*), user_interests(*, interests(*))").Eq("id", userID).Single().Execute(&user)
	if err != nil {
		return nil, err
	}

	profileResponse := &models.UserProfileResponse{
		UserID:          user.ID,
		Email:           user.Email,
		Username:        user.Username,
		Major:           user.Major,
		Gender:          user.Gender,
		NativeLanguage:  user.NativeLanguage,
		SpokenLanguages: user.SpokenLanguages,
		LearningLanguages: user.LearningLanguages,
		Residence:       user.Residence,
		Comment:         user.Comment,
		LastUpdated:     user.LastUpdatedAt,
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
		"major":             req.Major,
		"gender":            req.Gender,
		"native_language":   req.NativeLanguage,
		"spoken_languages":  req.SpokenLanguages,
		"learning_languages": req.LearningLanguages,
		"residence":         req.Residence,
		"comment":           req.Comment,
		"last_updated":      time.Now(),
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
	_, err = Supabase.DB.From("user_interests").Delete().Eq("user_id", userID).Execute(nil)
	if err != nil {
		return nil, err
	}

	// Then, insert new interests
	var userInterests []map[string]interface{}
	for _, interestID := range req.InterestIDs {
		userInterests = append(userInterests, map[string]interface{}{
			"user_id":    userID,
			"interest_id": interestID,
			"preference_level": 3, // Default preference level
		})
	}

	if len(userInterests) > 0 {
		_, err = Supabase.DB.From("user_interests").Insert(userInterests).Execute(nil)
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
		query = query.Filter("user_interests.interest_id", "eq", interestID)
	}

	if learningLanguage != "" {
		query = query.Filter("profiles.learning_languages", "cs", []string{learningLanguage})
	}

	if spokenLanguage != "" {
		query = query.Filter("profiles.spoken_languages", "cs", []string{spokenLanguage})
	}

	var users []models.User
	err := query.Execute(&users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func GetUserProfile(userID string) (*models.User, error) {
	var user models.User
	err := Supabase.DB.From("users").Select("id, username, profiles(major, gender, native_language, spoken_languages, learning_languages, residence, comment), user_interests(interests(id, name), preference_level)").Eq("id", userID).Single().Execute(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}