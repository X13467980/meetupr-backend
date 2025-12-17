package db

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
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
	// Only insert fields that exist in the users table
	userData := map[string]interface{}{
		"id":              user.ID,
		"email":           user.Email,
		"username":        user.Username,
		"is_oic_verified": user.IsOICVerified,
		"created_at":      user.CreatedAt,
	}
	var results []map[string]interface{}
	err := Supabase.DB.From("users").Insert(userData).Execute(&results)
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
		"avatar_url":         req.AvatarURL,
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
	log.Printf("GetUserChats: fetching chats for user %s", userID)

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

	// エラーハンドリング: 空の結果セットはエラーではない
	// Supabaseは結果が0件の場合、エラーではなく空の配列を返す
	// ただし、実際のエラー（ネットワークエラー、認証エラーなど）は処理する必要がある

	// エラーログを出力（デバッグ用）
	if err1 != nil {
		log.Printf("Error getting chats where user is user1 (userID=%s): %v", userID, err1)
		// Supabaseのエラーメッセージを確認
		if err1.Error() != "" {
			log.Printf("Error details: %s", err1.Error())
		}
	}
	if err2 != nil {
		log.Printf("Error getting chats where user is user2 (userID=%s): %v", userID, err2)
		// Supabaseのエラーメッセージを確認
		if err2.Error() != "" {
			log.Printf("Error details: %s", err2.Error())
		}
	}

	// 両方エラーの場合の処理
	// "unexpected end of JSON input"エラーは、空の結果セットの場合に発生する可能性がある
	if err1 != nil && err2 != nil {
		err1Str := err1.Error()
		err2Str := err2.Error()

		// "unexpected end of JSON input"は空の結果セットを示す可能性がある
		// この場合は空の配列を返す
		if containsIgnoreCase(err1Str, "unexpected end of json") &&
			containsIgnoreCase(err2Str, "unexpected end of json") {
			log.Printf("No chats found for user %s (empty result set)", userID)
			return []models.Chat{}, nil
		}

		// "not found"や空の結果を示すエラーの場合も空の配列を返す
		if (containsIgnoreCase(err1Str, "not found") || containsIgnoreCase(err1Str, "no rows")) &&
			(containsIgnoreCase(err2Str, "not found") || containsIgnoreCase(err2Str, "no rows")) {
			log.Printf("No chats found for user %s (both queries returned empty)", userID)
			return []models.Chat{}, nil
		}

		// その他のエラーは実際のエラーとして返す
		return nil, fmt.Errorf("failed to get chats: %v, %v", err1, err2)
	}

	log.Printf("GetUserChats: results1=%d, results2=%d", len(results1), len(results2))

	// エラーがない場合でも、結果が空の可能性がある
	// その場合は空の配列を返す（これは正常）
	if err1 == nil && err2 == nil && len(results1) == 0 && len(results2) == 0 {
		log.Printf("GetUserChats: no chats found for user %s (both queries returned empty arrays)", userID)
		return []models.Chat{}, nil
	}

	// Merge results and deduplicate by chat ID
	chatMap := make(map[int64]models.Chat)

	// Process results1
	if err1 == nil && len(results1) > 0 {
		log.Printf("GetUserChats: processing %d chat(s) where user is user1", len(results1))
		for _, result := range results1 {
			// 空のJSONをチェック
			if len(result) == 0 || string(result) == "null" {
				continue
			}
			var chat models.Chat
			if err := json.Unmarshal(result, &chat); err != nil {
				log.Printf("error unmarshalling chat: %v, raw: %s", err, string(result))
				continue
			}
			chatMap[chat.ID] = chat
		}
	}

	// Process results2
	if err2 == nil && len(results2) > 0 {
		log.Printf("GetUserChats: processing %d chat(s) where user is user2", len(results2))
		for _, result := range results2 {
			// 空のJSONをチェック
			if len(result) == 0 || string(result) == "null" {
				continue
			}
			var chat models.Chat
			if err := json.Unmarshal(result, &chat); err != nil {
				log.Printf("error unmarshalling chat: %v, raw: %s", err, string(result))
				continue
			}
			chatMap[chat.ID] = chat
		}
	}

	// Convert map to slice and populate additional info
	log.Printf("GetUserChats: found %d unique chat(s) for user %s", len(chatMap), userID)
	var chats []models.Chat
	for _, chat := range chatMap {
		// Determine the other user ID
		var otherUserID string
		if chat.User1ID == userID {
			otherUserID = chat.User2ID
		} else {
			otherUserID = chat.User1ID
		}

		// Get the other user's basic info (エラーが発生しても続行)
		if otherUserID != "" {
			otherUser, err := GetUserProfile(otherUserID)
			if err != nil {
				log.Printf("error getting other user profile for user %s: %v", otherUserID, err)
				// ユーザー情報が取得できなくてもチャットは返す
			} else {
				chat.OtherUser = otherUser
			}
		}

		// Get the last message (optimized: only fetch the last one)
		// エラーが発生しても続行
		lastMsg, err := GetLastChatMessage(chat.ID)
		if err != nil {
			log.Printf("error getting last message for chat %d: %v", chat.ID, err)
		} else if lastMsg != nil {
			chat.LastMessage = lastMsg
		}

		chats = append(chats, chat)
	}

	// Sort chats by created_at descending (most recent first)
	sort.Slice(chats, func(i, j int) bool {
		return chats[i].CreatedAt.After(chats[j].CreatedAt)
	})

	// チャットが存在しない場合は空のスライスを返す（エラーではない）
	if len(chats) == 0 {
		return []models.Chat{}, nil
	}

	return chats, nil
}

// containsIgnoreCase checks if a string contains a substring (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	return strings.Contains(s, substr)
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

// SearchUserResult represents a single user in search results
type SearchUserResult struct {
	UserID    string         `json:"user_id"`
	Username  string         `json:"username"`
	Comment   string         `json:"comment"`
	Residence string         `json:"residence"`
	AvatarURL string         `json:"avatar_url"`
	Interests []InterestItem `json:"interests"`
}

// InterestItem represents an interest/hobby item
type InterestItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// SearchUsersAdvanced searches for users with keyword, language, and country filters
// Excludes the current user from results
func SearchUsersAdvanced(currentUserID string, keyword string, languages []string, countries []string) ([]SearchUserResult, error) {
	log.Printf("SearchUsersAdvanced: currentUserID=%s, keyword=%s, languages=%v, countries=%v",
		currentUserID, keyword, languages, countries)

	// ユーザー情報とプロフィールを結合して取得
	var results []json.RawMessage
	err := Supabase.DB.From("users").
		Select("id, username, profiles(comment, residence, avatar_url, native_language, spoken_languages, learning_languages), user_interests(interests(id, name))").
		Neq("id", currentUserID).
		Execute(&results)
	if err != nil {
		log.Printf("SearchUsersAdvanced: error executing query: %v", err)
		return nil, err
	}

	log.Printf("SearchUsersAdvanced: found %d raw results", len(results))

	// 結果をパースしてフィルタリング
	var searchResults []SearchUserResult

	for _, raw := range results {
		var userData struct {
			ID       string `json:"id"`
			Username string `json:"username"`
			Profiles *struct {
				Comment           string   `json:"comment"`
				Residence         string   `json:"residence"`
				AvatarURL         string   `json:"avatar_url"`
				NativeLanguage    string   `json:"native_language"`
				SpokenLanguages   []string `json:"spoken_languages"`
				LearningLanguages []string `json:"learning_languages"`
			} `json:"profiles"`
			UserInterests []struct {
				Interests struct {
					ID   int    `json:"id"`
					Name string `json:"name"`
				} `json:"interests"`
			} `json:"user_interests"`
		}

		if err := json.Unmarshal(raw, &userData); err != nil {
			log.Printf("SearchUsersAdvanced: error unmarshalling user: %v", err)
			continue
		}

		// キーワードフィルタ（ユーザー名で部分一致）
		if keyword != "" {
			if !containsIgnoreCase(userData.Username, keyword) {
				continue
			}
		}

		// 言語フィルタ
		if len(languages) > 0 && userData.Profiles != nil {
			matched := false
			for _, lang := range languages {
				// native_language, spoken_languages, learning_languages のいずれかに一致
				if containsIgnoreCase(userData.Profiles.NativeLanguage, lang) {
					matched = true
					break
				}
				for _, spoken := range userData.Profiles.SpokenLanguages {
					if containsIgnoreCase(spoken, lang) {
						matched = true
						break
					}
				}
				if matched {
					break
				}
				for _, learning := range userData.Profiles.LearningLanguages {
					if containsIgnoreCase(learning, lang) {
						matched = true
						break
					}
				}
				if matched {
					break
				}
			}
			if !matched {
				continue
			}
		}

		// 国フィルタ（residence）
		if len(countries) > 0 && userData.Profiles != nil {
			matched := false
			for _, country := range countries {
				if containsIgnoreCase(userData.Profiles.Residence, country) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		// 結果を構築
		result := SearchUserResult{
			UserID:   userData.ID,
			Username: userData.Username,
		}

		if userData.Profiles != nil {
			result.Comment = userData.Profiles.Comment
			result.Residence = userData.Profiles.Residence
			result.AvatarURL = userData.Profiles.AvatarURL
		}

		// 趣味を追加
		for _, ui := range userData.UserInterests {
			result.Interests = append(result.Interests, InterestItem{
				ID:   ui.Interests.ID,
				Name: ui.Interests.Name,
			})
		}

		searchResults = append(searchResults, result)
	}

	log.Printf("SearchUsersAdvanced: returning %d filtered results", len(searchResults))

	// 結果が nil の場合は空の配列を返す
	if searchResults == nil {
		return []SearchUserResult{}, nil
	}

	return searchResults, nil
}

// IsChatParticipant checks if a user is a participant in a chat room
func IsChatParticipant(chatID int64, userID string) (bool, error) {
	log.Printf("IsChatParticipant: checking if user %s is participant in chat %d", userID, chatID)

	// First, get the chat
	var results []json.RawMessage
	err := Supabase.DB.From("chats").
		Select("id, user1_id, user2_id").
		Eq("id", strconv.FormatInt(chatID, 10)).
		Execute(&results)

	if err != nil {
		errStr := err.Error()
		log.Printf("IsChatParticipant: error querying chat %d: %v", chatID, err)

		// "unexpected end of JSON input"は空の結果セットを示す可能性がある
		if containsIgnoreCase(errStr, "unexpected end of json") {
			log.Printf("IsChatParticipant: chat %d not found (empty result set)", chatID)
			return false, nil
		}

		// "not found"や空の結果を示すエラーの場合もfalseを返す
		if containsIgnoreCase(errStr, "not found") || containsIgnoreCase(errStr, "no rows") {
			log.Printf("IsChatParticipant: chat %d not found", chatID)
			return false, nil
		}

		return false, err
	}

	if len(results) == 0 {
		log.Printf("IsChatParticipant: chat %d not found (no results)", chatID)
		return false, nil
	}

	// Check if user is either user1 or user2
	var chat struct {
		ID      int64  `json:"id"`
		User1ID string `json:"user1_id"`
		User2ID string `json:"user2_id"`
	}

	// Check if result is empty or null before unmarshalling
	if len(results[0]) == 0 || string(results[0]) == "null" {
		log.Printf("IsChatParticipant: chat %d result is empty or null", chatID)
		return false, nil
	}

	if err := json.Unmarshal(results[0], &chat); err != nil {
		log.Printf("IsChatParticipant: error unmarshalling chat %d: %v", chatID, err)
		return false, err
	}

	isParticipant := chat.User1ID == userID || chat.User2ID == userID
	log.Printf("IsChatParticipant: user %s is participant in chat %d: %v", userID, chatID, isParticipant)
	return isParticipant, nil
}
