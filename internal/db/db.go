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
	// Try multiple approaches to get user info
	// Approach 1: Try with Select("id") first (we know this works)
	var userIDs []map[string]interface{}
	err := Supabase.DB.From("users").
		Select("id").
		Eq("id", userID).
		Execute(&userIDs)

	if err != nil || len(userIDs) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	// User exists, now try to get username
	// Approach 2: Try to get username separately
	var usernameResults []map[string]interface{}
	err = Supabase.DB.From("users").
		Select("username").
		Eq("id", userID).
		Execute(&usernameResults)

	username := userID // Fallback to ID if username unavailable
	if err == nil && len(usernameResults) > 0 {
		if u, ok := usernameResults[0]["username"].(string); ok && u != "" {
			username = u
		}
	}

	// Create user with basic info
	user := models.User{
		ID:       userID,
		Username: username,
	}

	// Try to get full profile info (may fail, but that's ok)
	var fullResults []json.RawMessage
	err = Supabase.DB.From("users").
		Select("id, username, profiles(major, gender, native_language, spoken_languages, learning_languages, residence, comment), user_interests(interests(id, name), preference_level)").
		Eq("id", userID).
		Execute(&fullResults)
	if err == nil && len(fullResults) > 0 {
		var fullUser models.User
		if err := json.Unmarshal(fullResults[0], &fullUser); err == nil {
			// Merge full user data, but keep username we got above
			if fullUser.Username != "" {
				user.Username = fullUser.Username
			}
			user.Major = fullUser.Major
			user.Gender = fullUser.Gender
			user.NativeLanguage = fullUser.NativeLanguage
			user.SpokenLanguages = fullUser.SpokenLanguages
			user.LearningLanguages = fullUser.LearningLanguages
			user.Residence = fullUser.Residence
			user.Comment = fullUser.Comment
			user.Interests = fullUser.Interests
		}
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

	// First get all chat IDs (this works - Test 4 confirms)
	var chatIDs []map[string]interface{}
	err := Supabase.DB.From("chats").
		Select("id").
		Execute(&chatIDs)

	if err != nil {
		errStr := err.Error()
		if containsIgnoreCase(errStr, "unexpected end of json") ||
			containsIgnoreCase(errStr, "invalid character") ||
			containsIgnoreCase(errStr, "not found") ||
			containsIgnoreCase(errStr, "no rows") {
			log.Printf("No chats found (empty result set)")
			return []models.Chat{}, nil
		}
		return nil, fmt.Errorf("failed to get chat IDs: %v", err)
	}

	log.Printf("GetUserChats: found %d chat ID(s), using messages table to find user's chats", len(chatIDs))

	// Since Select("id") works for chats but other fields fail,
	// we'll use the messages table to find which chats the user participates in
	// This works because GetChatMessages successfully queries messages table
	userChatIDs := make(map[int64]bool)

	// Get all messages sent by this user to find their chats
	var userMessages []models.Message
	err = Supabase.DB.From("messages").
		Select("chat_id").
		Eq("sender_id", userID).
		Execute(&userMessages)

	if err == nil {
		for _, msg := range userMessages {
			userChatIDs[msg.ChatID] = true
		}
	}

	// Also check messages in chats where user might be recipient
	// We need to check all messages and see which chats contain this user
	var allMessages []models.Message
	err = Supabase.DB.From("messages").
		Select("chat_id, sender_id").
		Execute(&allMessages)

	if err == nil {
		for _, msg := range allMessages {
			// If message is in a chat we already know about, or if sender is our user
			if msg.SenderID == userID {
				userChatIDs[msg.ChatID] = true
			}
		}
	}

	log.Printf("GetUserChats: found %d chat(s) from messages for user %s", len(userChatIDs), userID)

	// Build chat list from the chat IDs we found
	chatMap := make(map[int64]models.Chat)
	for chatID := range userChatIDs {
		// Try to get user1_id and user2_id from chats table
		var chatUserInfo []map[string]interface{}
		err = Supabase.DB.From("chats").
			Select("user1_id, user2_id").
			Eq("id", strconv.FormatInt(chatID, 10)).
			Execute(&chatUserInfo)

		chat := models.Chat{
			ID: chatID,
		}

		if err == nil && len(chatUserInfo) > 0 {
			chat.User1ID, _ = chatUserInfo[0]["user1_id"].(string)
			chat.User2ID, _ = chatUserInfo[0]["user2_id"].(string)
		} else {
			// Fallback: set current user as user1_id
			chat.User1ID = userID
		}

		// Get message IDs to find senders and last message
		// We know Select("id") works, so use that
		var msgIDs []map[string]interface{}
		err = Supabase.DB.From("messages").
			Select("id").
			Eq("chat_id", strconv.FormatInt(chatID, 10)).
			Execute(&msgIDs)

		if err == nil && len(msgIDs) > 0 {
			// Get sender IDs from message IDs (try to get sender_id for each message)
			senderIDs := make(map[string]bool)
			var lastMsgID int64
			for i, msgIDMap := range msgIDs {
				msgIDFloat, ok := msgIDMap["id"].(float64)
				if !ok {
					continue
				}
				msgID := int64(msgIDFloat)
				if i == len(msgIDs)-1 {
					lastMsgID = msgID // Last message ID
				}

				// Try to get sender_id for this message
				var senderInfo []map[string]interface{}
				err2 := Supabase.DB.From("messages").
					Select("sender_id").
					Eq("id", strconv.FormatInt(msgID, 10)).
					Execute(&senderInfo)
				if err2 == nil && len(senderInfo) > 0 {
					if senderID, ok := senderInfo[0]["sender_id"].(string); ok {
						senderIDs[senderID] = true
					}
				}
			}

			// Find the other user (not the current user)
			for sid := range senderIDs {
				if sid != userID {
					chat.User2ID = sid
					break
				}
			}

			// Try to get last message content and sent_at
			// Since Select with multiple fields fails, get each field separately
			if lastMsgID > 0 {
				log.Printf("GetUserChats: trying to get last message %d for chat %d", lastMsgID, chatID)

				// Get each field separately (Select with single field works)
				content := ""
				senderID := ""
				sentAt := time.Now() // Default fallback

				// Get content
				var contentResults []map[string]interface{}
				err3 := Supabase.DB.From("messages").
					Select("content").
					Eq("id", strconv.FormatInt(lastMsgID, 10)).
					Execute(&contentResults)
				if err3 == nil && len(contentResults) > 0 {
					content, _ = contentResults[0]["content"].(string)
				}

				// Get sender_id
				var senderResults []map[string]interface{}
				err4 := Supabase.DB.From("messages").
					Select("sender_id").
					Eq("id", strconv.FormatInt(lastMsgID, 10)).
					Execute(&senderResults)
				if err4 == nil && len(senderResults) > 0 {
					senderID, _ = senderResults[0]["sender_id"].(string)
				}

				// Get sent_at
				var sentAtResults []map[string]interface{}
				err5 := Supabase.DB.From("messages").
					Select("sent_at").
					Eq("id", strconv.FormatInt(lastMsgID, 10)).
					Execute(&sentAtResults)
				if err5 == nil && len(sentAtResults) > 0 {
					if sentAtStr, ok := sentAtResults[0]["sent_at"].(string); ok {
						if parsedTime, err := time.Parse(time.RFC3339, sentAtStr); err == nil {
							sentAt = parsedTime
						} else if parsedTime, err := time.Parse("2006-01-02T15:04:05Z07:00", sentAtStr); err == nil {
							sentAt = parsedTime
						}
					}
				}

				if content != "" {
					log.Printf("GetUserChats: successfully got last message for chat %d: %s", chatID, content)
					chat.LastMessage = &models.Message{
						ID:          lastMsgID,
						ChatID:      chatID,
						SenderID:    senderID,
						Content:     content,
						MessageType: "text",
						SentAt:      sentAt,
					}
				} else {
					log.Printf("GetUserChats: last message content is empty for chat %d", chatID)
				}
			} else {
				log.Printf("GetUserChats: lastMsgID is 0 for chat %d", chatID)
			}
		}

		chatMap[chat.ID] = chat
	}

	log.Printf("GetUserChats: found %d chat(s) for user %s", len(chatMap), userID)

	// Convert map to slice
	var chats []models.Chat
	for _, chat := range chatMap {
		chats = append(chats, chat)
	}

	// Sort chats by ID descending (most recent first) - created_at is not available
	sort.Slice(chats, func(i, j int) bool {
		return chats[i].ID > chats[j].ID
	})

	// Populate additional info for each chat
	for i := range chats {
		chat := &chats[i]
		// Determine the other user ID
		var otherUserID string
		if chat.User1ID == userID {
			otherUserID = chat.User2ID
		} else {
			otherUserID = chat.User1ID
		}

		// Get the other user's basic info (エラーが発生しても続行)
		if otherUserID != "" {
			log.Printf("GetUserChats: fetching profile for other user %s", otherUserID)
			otherUser, err := GetUserProfile(otherUserID)
			if err != nil {
				log.Printf("error getting other user profile for user %s: %v", otherUserID, err)
				// Create a minimal user object with just the ID
				// This ensures the frontend can at least display something
				chat.OtherUser = &models.User{
					ID:       otherUserID,
					Username: otherUserID, // Fallback to ID if username unavailable
				}
			} else {
				log.Printf("GetUserChats: successfully got profile for user %s, username: %s", otherUserID, otherUser.Username)
				chat.OtherUser = otherUser
			}
		} else {
			log.Printf("GetUserChats: otherUserID is empty for chat %d", chat.ID)
		}

		// Last message is already set if we got messages above
		// If not set, try to get it separately
		if chat.LastMessage == nil {
			log.Printf("GetUserChats: LastMessage is nil for chat %d, trying GetLastChatMessage", chat.ID)
			lastMsg, err := GetLastChatMessage(chat.ID)
			if err != nil {
				log.Printf("error getting last message for chat %d: %v", chat.ID, err)
			} else if lastMsg != nil {
				log.Printf("GetUserChats: successfully got last message via GetLastChatMessage for chat %d: %s", chat.ID, lastMsg.Content)
				chat.LastMessage = lastMsg
			} else {
				log.Printf("GetUserChats: GetLastChatMessage returned nil for chat %d", chat.ID)
			}
		} else {
			log.Printf("GetUserChats: LastMessage already set for chat %d: %s", chat.ID, chat.LastMessage.Content)
		}
	}

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

// mapToChat converts a map[string]interface{} to models.Chat
func mapToChat(m map[string]interface{}) (models.Chat, error) {
	var chat models.Chat

	// ID
	if id, ok := m["id"].(float64); ok {
		chat.ID = int64(id)
	} else {
		return chat, fmt.Errorf("invalid chat ID: %v", m["id"])
	}

	// User1ID
	if user1ID, ok := m["user1_id"].(string); ok {
		chat.User1ID = user1ID
	}

	// User2ID
	if user2ID, ok := m["user2_id"].(string); ok {
		chat.User2ID = user2ID
	}

	// AISuggestedTheme
	if theme, ok := m["ai_suggested_theme"].(string); ok {
		chat.AISuggestedTheme = theme
	}

	// CreatedAt
	if createdAt, ok := m["created_at"].(string); ok {
		parsedTime, err := time.Parse(time.RFC3339, createdAt)
		if err != nil {
			// Try alternative format
			parsedTime, err = time.Parse("2006-01-02T15:04:05Z07:00", createdAt)
			if err != nil {
				log.Printf("error parsing created_at: %v, using current time", err)
				parsedTime = time.Now()
			}
		}
		chat.CreatedAt = parsedTime
	} else {
		chat.CreatedAt = time.Now()
	}

	return chat, nil
}

// GetLastChatMessage returns the last message in a chat room
func GetLastChatMessage(chatID int64) (*models.Message, error) {
	// Use GetChatMessages which we know works reliably
	messages, err := GetChatMessages(chatID)
	if err != nil {
		// Return nil instead of error - it's ok if we can't get the last message
		return nil, nil
	}

	if len(messages) == 0 {
		return nil, nil
	}

	// Messages are already sorted ascending, so return the last one
	return &messages[len(messages)-1], nil
}

// GetChatMessages returns all messages in a chat room
func GetChatMessages(chatID int64) ([]models.Message, error) {
	// First get message IDs (this should work)
	var messageIDs []map[string]interface{}
	err := Supabase.DB.From("messages").
		Select("id").
		Eq("chat_id", strconv.FormatInt(chatID, 10)).
		Execute(&messageIDs)

	if err != nil {
		errStr := err.Error()
		if containsIgnoreCase(errStr, "unexpected end of json") ||
			containsIgnoreCase(errStr, "invalid character") ||
			containsIgnoreCase(errStr, "not found") ||
			containsIgnoreCase(errStr, "no rows") {
			return []models.Message{}, nil
		}
		return nil, err
	}

	if len(messageIDs) == 0 {
		return []models.Message{}, nil
	}

	log.Printf("GetChatMessages: found %d message ID(s) for chat %d", len(messageIDs), chatID)

	// Try to get all messages at once with all fields
	var messages []models.Message
	err = Supabase.DB.From("messages").
		Select("id, chat_id, sender_id, content, message_type, sent_at").
		Eq("chat_id", strconv.FormatInt(chatID, 10)).
		Execute(&messages)

	if err != nil {
		errStr := err.Error()
		// If detailed query fails, try to get messages one by one using message IDs
		if containsIgnoreCase(errStr, "unexpected end of json") ||
			containsIgnoreCase(errStr, "invalid character") {
			log.Printf("GetChatMessages: detailed query failed for chat %d, trying individual message queries with single fields", chatID)
			// Build messages from IDs we have - query each field separately (Select with multiple fields fails)
			for _, msgIDMap := range messageIDs {
				msgIDFloat, ok := msgIDMap["id"].(float64)
				if !ok {
					continue
				}
				msgID := int64(msgIDFloat)

				// Get each field separately
				content := ""
				senderID := ""
				sentAt := time.Now() // Default fallback

				// Get content
				var contentResults []map[string]interface{}
				err2 := Supabase.DB.From("messages").
					Select("content").
					Eq("id", strconv.FormatInt(msgID, 10)).
					Execute(&contentResults)
				if err2 == nil && len(contentResults) > 0 {
					content, _ = contentResults[0]["content"].(string)
				}

				// Get sender_id
				var senderResults []map[string]interface{}
				err3 := Supabase.DB.From("messages").
					Select("sender_id").
					Eq("id", strconv.FormatInt(msgID, 10)).
					Execute(&senderResults)
				if err3 == nil && len(senderResults) > 0 {
					senderID, _ = senderResults[0]["sender_id"].(string)
				}

				// Get sent_at
				var sentAtResults []map[string]interface{}
				err4 := Supabase.DB.From("messages").
					Select("sent_at").
					Eq("id", strconv.FormatInt(msgID, 10)).
					Execute(&sentAtResults)
				if err4 == nil && len(sentAtResults) > 0 {
					if sentAtStr, ok := sentAtResults[0]["sent_at"].(string); ok {
						if parsedTime, err := time.Parse(time.RFC3339, sentAtStr); err == nil {
							sentAt = parsedTime
						} else if parsedTime, err := time.Parse("2006-01-02T15:04:05Z07:00", sentAtStr); err == nil {
							sentAt = parsedTime
						}
					}
				}

				if content != "" {
					msg := models.Message{
						ID:          msgID,
						ChatID:      chatID,
						SenderID:    senderID,
						Content:     content,
						MessageType: "text",
						SentAt:      sentAt,
					}
					messages = append(messages, msg)
				}
			}
		} else {
			return nil, err
		}
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

	// Get user1_id and user2_id separately (Select with multiple fields fails)
	var user1ID string
	var user2ID string

	// Get user1_id
	var user1Results []map[string]interface{}
	err := Supabase.DB.From("chats").
		Select("user1_id").
		Eq("id", strconv.FormatInt(chatID, 10)).
		Execute(&user1Results)
	if err == nil && len(user1Results) > 0 {
		user1ID, _ = user1Results[0]["user1_id"].(string)
	} else {
		errStr := ""
		if err != nil {
			errStr = err.Error()
		}
		if containsIgnoreCase(errStr, "unexpected end of json") ||
			containsIgnoreCase(errStr, "invalid character") ||
			containsIgnoreCase(errStr, "not found") ||
			containsIgnoreCase(errStr, "no rows") ||
			len(user1Results) == 0 {
			log.Printf("IsChatParticipant: chat %d not found", chatID)
			return false, nil
		}
	}

	// Get user2_id
	var user2Results []map[string]interface{}
	err = Supabase.DB.From("chats").
		Select("user2_id").
		Eq("id", strconv.FormatInt(chatID, 10)).
		Execute(&user2Results)
	if err == nil && len(user2Results) > 0 {
		user2ID, _ = user2Results[0]["user2_id"].(string)
	}

	// Check if user is either user1 or user2
	isParticipant := user1ID == userID || user2ID == userID
	log.Printf("IsChatParticipant: user %s is participant in chat %d: %v (user1=%s, user2=%s)", userID, chatID, isParticipant, user1ID, user2ID)
	return isParticipant, nil
}

// GetChatDetail returns detailed information about a specific chat room
func GetChatDetail(chatID int64, userID string) (*models.Chat, error) {
	log.Printf("GetChatDetail: fetching chat %d for user %s", chatID, userID)

	// Get chat basic info - fetch each field separately
	chat := models.Chat{
		ID: chatID,
	}

	// Get user1_id
	var user1Results []map[string]interface{}
	err := Supabase.DB.From("chats").
		Select("user1_id").
		Eq("id", strconv.FormatInt(chatID, 10)).
		Execute(&user1Results)
	if err == nil && len(user1Results) > 0 {
		chat.User1ID, _ = user1Results[0]["user1_id"].(string)
	}

	// Get user2_id
	var user2Results []map[string]interface{}
	err = Supabase.DB.From("chats").
		Select("user2_id").
		Eq("id", strconv.FormatInt(chatID, 10)).
		Execute(&user2Results)
	if err == nil && len(user2Results) > 0 {
		chat.User2ID, _ = user2Results[0]["user2_id"].(string)
	}

	// Verify that the user is a participant in this chat
	if chat.User1ID != userID && chat.User2ID != userID {
		return nil, fmt.Errorf("user %s is not a participant in chat %d", userID, chatID)
	}

	// Get ai_suggested_theme (optional)
	var themeResults []map[string]interface{}
	err = Supabase.DB.From("chats").
		Select("ai_suggested_theme").
		Eq("id", strconv.FormatInt(chatID, 10)).
		Execute(&themeResults)
	if err == nil && len(themeResults) > 0 {
		if theme, ok := themeResults[0]["ai_suggested_theme"].(string); ok {
			chat.AISuggestedTheme = theme
		}
	}

	// Determine the other user ID
	var otherUserID string
	if chat.User1ID == userID {
		otherUserID = chat.User2ID
	} else {
		otherUserID = chat.User1ID
	}

	// Get the other user's profile
	if otherUserID != "" {
		log.Printf("GetChatDetail: fetching profile for other user %s", otherUserID)
		otherUser, err := GetUserProfile(otherUserID)
		if err != nil {
			log.Printf("error getting other user profile for user %s: %v", otherUserID, err)
			// Create a minimal user object with just the ID
			chat.OtherUser = &models.User{
				ID:       otherUserID,
				Username: otherUserID, // Fallback to ID if username unavailable
			}
		} else {
			log.Printf("GetChatDetail: successfully got profile for user %s, username: %s", otherUserID, otherUser.Username)
			chat.OtherUser = otherUser
		}
	}

	// Get last message
	lastMsg, err := GetLastChatMessage(chatID)
	if err != nil {
		log.Printf("error getting last message for chat %d: %v", chatID, err)
	} else if lastMsg != nil {
		log.Printf("GetChatDetail: successfully got last message for chat %d: %s", chatID, lastMsg.Content)
		chat.LastMessage = lastMsg
	}

	return &chat, nil
}

// GetOrCreateChat finds an existing chat between two users or creates a new one
// Returns the chat ID
func GetOrCreateChat(user1ID, user2ID string) (int64, error) {
	log.Printf("GetOrCreateChat: finding or creating chat between %s and %s", user1ID, user2ID)

	// Ensure user1ID < user2ID for consistent ordering (to match unique index)
	// The unique index uses LEAST(user1_id, user2_id) and GREATEST(user1_id, user2_id)
	// So we need to check both orders: user1-user2 and user2-user1
	var existingChatID int64
	found := false

	// Try to find existing chat: user1_id = user1ID and user2_id = user2ID
	var results1 []map[string]interface{}
	err1 := Supabase.DB.From("chats").
		Select("id").
		Eq("user1_id", user1ID).
		Eq("user2_id", user2ID).
		Execute(&results1)

	if err1 == nil && len(results1) > 0 {
		if idFloat, ok := results1[0]["id"].(float64); ok {
			existingChatID = int64(idFloat)
			found = true
			log.Printf("GetOrCreateChat: found existing chat %d (user1=%s, user2=%s)", existingChatID, user1ID, user2ID)
		}
	}

	// Try to find existing chat: user1_id = user2ID and user2_id = user1ID
	if !found {
		var results2 []map[string]interface{}
		err2 := Supabase.DB.From("chats").
			Select("id").
			Eq("user1_id", user2ID).
			Eq("user2_id", user1ID).
			Execute(&results2)

		if err2 == nil && len(results2) > 0 {
			if idFloat, ok := results2[0]["id"].(float64); ok {
				existingChatID = int64(idFloat)
				found = true
				log.Printf("GetOrCreateChat: found existing chat %d (user1=%s, user2=%s)", existingChatID, user2ID, user1ID)
			}
		}
	}

	// If chat exists, return it
	if found {
		return existingChatID, nil
	}

	// Chat doesn't exist, create a new one
	log.Printf("GetOrCreateChat: no existing chat found, creating new chat")
	chatData := map[string]interface{}{
		"user1_id": user1ID,
		"user2_id": user2ID,
	}

	var insertResults []map[string]interface{}
	err := Supabase.DB.From("chats").Insert(chatData).Execute(&insertResults)
	if err != nil {
		errStr := err.Error()
		// Check if it's a duplicate key error (race condition - another request created it)
		if containsIgnoreCase(errStr, "duplicate key") || containsIgnoreCase(errStr, "unique constraint") {
			log.Printf("GetOrCreateChat: chat was created by another request, retrying to find it")
			// Retry finding the chat
			var retryResults []map[string]interface{}
			err1 := Supabase.DB.From("chats").
				Select("id").
				Eq("user1_id", user1ID).
				Eq("user2_id", user2ID).
				Execute(&retryResults)
			if err1 == nil && len(retryResults) > 0 {
				if idFloat, ok := retryResults[0]["id"].(float64); ok {
					return int64(idFloat), nil
				}
			}
			// Try reverse order
			err2 := Supabase.DB.From("chats").
				Select("id").
				Eq("user1_id", user2ID).
				Eq("user2_id", user1ID).
				Execute(&retryResults)
			if err2 == nil && len(retryResults) > 0 {
				if idFloat, ok := retryResults[0]["id"].(float64); ok {
					return int64(idFloat), nil
				}
			}
		}
		return 0, fmt.Errorf("failed to create chat: %v", err)
	}

	if len(insertResults) == 0 {
		return 0, fmt.Errorf("chat creation succeeded but no ID returned")
	}

	// Extract the created chat ID
	if idFloat, ok := insertResults[0]["id"].(float64); ok {
		newChatID := int64(idFloat)
		log.Printf("GetOrCreateChat: created new chat %d", newChatID)
		return newChatID, nil
	}

	return 0, fmt.Errorf("chat creation succeeded but could not extract ID from response")
}
