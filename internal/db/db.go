package db

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
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

	// Get avatar_url separately (to avoid JSON parsing issues with nested profiles)
	var avatarResults []map[string]interface{}
	err = Supabase.DB.From("profiles").
		Select("avatar_url").
		Eq("user_id", userID).
		Execute(&avatarResults)
	if err == nil && len(avatarResults) > 0 {
		if avatarURL, ok := avatarResults[0]["avatar_url"].(string); ok && avatarURL != "" {
			user.AvatarURL = avatarURL
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

	// Optimized: Get chat IDs directly from chats table where user is participant
	// Since we can't use OR in supabase-go, we'll get chats where user1_id = userID, then user2_id = userID
	var chatIDs1 []map[string]interface{}
	err1 := Supabase.DB.From("chats").
		Select("id").
		Eq("user1_id", userID).
		Execute(&chatIDs1)

	var chatIDs2 []map[string]interface{}
	err2 := Supabase.DB.From("chats").
		Select("id").
		Eq("user2_id", userID).
		Execute(&chatIDs2)

	// Combine chat IDs
	userChatIDs := make(map[int64]bool)
	if err1 == nil {
		for _, chatIDMap := range chatIDs1 {
			if idFloat, ok := chatIDMap["id"].(float64); ok {
				userChatIDs[int64(idFloat)] = true
			}
		}
	}
	if err2 == nil {
		for _, chatIDMap := range chatIDs2 {
			if idFloat, ok := chatIDMap["id"].(float64); ok {
				userChatIDs[int64(idFloat)] = true
			}
		}
	}

	log.Printf("GetUserChats: found %d chat(s) for user %s", len(userChatIDs), userID)

	if len(userChatIDs) == 0 {
		return []models.Chat{}, nil
	}

	// Build chat list from the chat IDs we found
	// Try to get user1_id and user2_id together first (may fail due to supabase-go issue)
	chatMap := make(map[int64]models.Chat)
	for chatID := range userChatIDs {
		chat := models.Chat{
			ID: chatID,
		}

		// Try to get both user1_id and user2_id together
		var chatUserInfo []map[string]interface{}
		err := Supabase.DB.From("chats").
			Select("user1_id, user2_id").
			Eq("id", strconv.FormatInt(chatID, 10)).
			Execute(&chatUserInfo)

		if err == nil && len(chatUserInfo) > 0 {
			// Successfully got both fields together
			chat.User1ID, _ = chatUserInfo[0]["user1_id"].(string)
			chat.User2ID, _ = chatUserInfo[0]["user2_id"].(string)
		} else {
			// Fallback: get each field separately
			var user1Results []map[string]interface{}
			err1 := Supabase.DB.From("chats").
				Select("user1_id").
				Eq("id", strconv.FormatInt(chatID, 10)).
				Execute(&user1Results)
			if err1 == nil && len(user1Results) > 0 {
				chat.User1ID, _ = user1Results[0]["user1_id"].(string)
			}

			var user2Results []map[string]interface{}
			err2 := Supabase.DB.From("chats").
				Select("user2_id").
				Eq("id", strconv.FormatInt(chatID, 10)).
				Execute(&user2Results)
			if err2 == nil && len(user2Results) > 0 {
				chat.User2ID, _ = user2Results[0]["user2_id"].(string)
			}
		}

		chatMap[chat.ID] = chat
	}

	log.Printf("GetUserChats: built %d chat(s) for user %s", len(chatMap), userID)

	// Convert map to slice
	var chats []models.Chat
	for _, chat := range chatMap {
		chats = append(chats, chat)
	}

	// Sort chats by ID descending (most recent first) - created_at is not available
	sort.Slice(chats, func(i, j int) bool {
		return chats[i].ID > chats[j].ID
	})

	// Populate additional info for each chat (optimized: use GetLastChatMessage directly)
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
			otherUser, err := GetUserProfile(otherUserID)
			if err != nil {
				log.Printf("error getting other user profile for user %s: %v", otherUserID, err)
				// Create a minimal user object with just the ID
				chat.OtherUser = &models.User{
					ID:       otherUserID,
					Username: otherUserID, // Fallback to ID if username unavailable
				}
			} else {
				chat.OtherUser = otherUser
			}
		}

		// Get last message using optimized GetLastChatMessage function
		lastMsg, err := GetLastChatMessage(chat.ID)
		if err != nil {
			log.Printf("error getting last message for chat %d: %v", chat.ID, err)
		} else if lastMsg != nil {
			chat.LastMessage = lastMsg
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

// GetLastChatMessage returns the last message in a chat room (optimized to only get the last message)
func GetLastChatMessage(chatID int64) (*models.Message, error) {
	// Get the last message ID by getting all message IDs and taking the last one
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
			return nil, nil
		}
		return nil, err
	}

	if len(messageIDs) == 0 {
		return nil, nil
	}

	// Get the last message ID
	var lastMsgID int64
	if idFloat, ok := messageIDs[len(messageIDs)-1]["id"].(float64); ok {
		lastMsgID = int64(idFloat)
	} else {
		return nil, nil
	}

	// Get each field separately for the last message
	content := ""
	senderID := ""
	sentAt := time.Now()

	// Get content
	var contentResults []map[string]interface{}
	err = Supabase.DB.From("messages").
		Select("content").
		Eq("id", strconv.FormatInt(lastMsgID, 10)).
		Execute(&contentResults)
	if err == nil && len(contentResults) > 0 {
		content, _ = contentResults[0]["content"].(string)
	}

	// Get sender_id
	var senderResults []map[string]interface{}
	err = Supabase.DB.From("messages").
		Select("sender_id").
		Eq("id", strconv.FormatInt(lastMsgID, 10)).
		Execute(&senderResults)
	if err == nil && len(senderResults) > 0 {
		senderID, _ = senderResults[0]["sender_id"].(string)
	}

	// Get sent_at
	var sentAtResults []map[string]interface{}
	err = Supabase.DB.From("messages").
		Select("sent_at").
		Eq("id", strconv.FormatInt(lastMsgID, 10)).
		Execute(&sentAtResults)
	if err == nil && len(sentAtResults) > 0 {
		if sentAtStr, ok := sentAtResults[0]["sent_at"].(string); ok {
			if parsedTime, err := time.Parse(time.RFC3339, sentAtStr); err == nil {
				sentAt = parsedTime
			} else if parsedTime, err := time.Parse("2006-01-02T15:04:05Z07:00", sentAtStr); err == nil {
				sentAt = parsedTime
			}
		}
	}

	if content == "" {
		return nil, nil
	}

	return &models.Message{
		ID:          lastMsgID,
		ChatID:      chatID,
		SenderID:    senderID,
		Content:     content,
		MessageType: "text",
		SentAt:      sentAt,
	}, nil
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
	AvatarURL *string        `json:"avatar_url"` // NULLの可能性があるためポインタ型
	Interests []InterestItem `json:"interests"`
}

// InterestItem represents an interest/hobby item
type InterestItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// SearchUsersAdvanced filters users based on keyword, languages, and countries
// Returns users (excluding current user) whose profile matches the specified criteria:
//   - keyword: partial match in username (optional)
//   - languages: users who have any of the specified languages in their profile
//     (checks native_language, spoken_languages, or learning_languages)
//   - countries: users who have any of the specified countries in their residence field
//
// Excludes the current user from results
func SearchUsersAdvanced(currentUserID string, keyword string, languages []string, countries []string) ([]SearchUserResult, error) {
	log.Printf("SearchUsersAdvanced: currentUserID=%s, keyword=%s, languages=%v, countries=%v",
		currentUserID, keyword, languages, countries)

	// まず、ユーザーIDだけを取得（これは動作する）
	var userIDs []map[string]interface{}
	err := Supabase.DB.From("users").
		Select("id").
		Neq("id", currentUserID).
		Execute(&userIDs)
	if err != nil {
		errStr := err.Error()
		log.Printf("SearchUsersAdvanced: error executing basic query: %v", err)

		// unexpected end of JSON input エラーの場合は空の結果を返す
		if containsIgnoreCase(errStr, "unexpected end of json") ||
			containsIgnoreCase(errStr, "invalid character") ||
			containsIgnoreCase(errStr, "not found") ||
			containsIgnoreCase(errStr, "no rows") {
			log.Printf("SearchUsersAdvanced: query returned empty or invalid result, returning empty array")
			return []SearchUserResult{}, nil
		}

		return nil, fmt.Errorf("failed to search users: %v", err)
	}

	log.Printf("SearchUsersAdvanced: found %d user IDs", len(userIDs))

	// フィルター条件の有無を確認
	hasFilters := keyword != "" || len(languages) > 0 || len(countries) > 0
	log.Printf("SearchUsersAdvanced: hasFilters=%v (keyword=%q, languages=%v, countries=%v)",
		hasFilters, keyword, languages, countries)

	// 結果をパースしてフィルタリング（並列処理で高速化）
	var searchResults []SearchUserResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 並列処理の最大数（同時実行数を制限してDBへの負荷を軽減）
	maxWorkers := 10
	semaphore := make(chan struct{}, maxWorkers)

	for _, userIDMap := range userIDs {
		wg.Add(1)
		semaphore <- struct{}{} // セマフォを取得

		go func(userIDMap map[string]interface{}) {
			defer wg.Done()
			defer func() { <-semaphore }() // セマフォを解放

			// エラーが発生した場合はログを出力してスキップ
			defer func() {
				if r := recover(); r != nil {
					log.Printf("SearchUsersAdvanced: panic in goroutine: %v", r)
				}
			}()

			// IDを取得（string型として扱う）
			userID := ""
			if idStr, ok := userIDMap["id"].(string); ok {
				userID = idStr
			} else if idFloat, ok := userIDMap["id"].(float64); ok {
				// float64の場合は文字列に変換を試みる（通常は発生しないが念のため）
				userID = strconv.FormatFloat(idFloat, 'f', -1, 64)
			}

			if userID == "" {
				return
			}

			// usernameを個別に取得
			var usernameResults []map[string]interface{}
			err2 := Supabase.DB.From("users").
				Select("username").
				Eq("id", userID).
				Execute(&usernameResults)

			// エラーハンドリング: unexpected end of JSON input エラーは無視して続行
			if err2 != nil {
				errStr := err2.Error()
				if !containsIgnoreCase(errStr, "unexpected end of json") &&
					!containsIgnoreCase(errStr, "invalid character") {
					log.Printf("SearchUsersAdvanced: error getting username for user %s: %v", userID, err2)
				}
			}

			username := userID // フォールバック
			if err2 == nil && len(usernameResults) > 0 {
				if u, ok := usernameResults[0]["username"].(string); ok && u != "" {
					username = u
				}
			}

			// キーワードフィルタ（ユーザー名で部分一致）
			if keyword != "" {
				if !containsIgnoreCase(username, keyword) {
					return
				}
			}

			// 言語または国のフィルタがある場合、プロフィール情報を取得する必要がある
			needProfileInfo := len(languages) > 0 || len(countries) > 0

			profileData := make(map[string]interface{})
			if needProfileInfo {
				// プロフィール情報を個別フィールドで取得（複数フィールドの同時取得が失敗するため）
				// フィルタリングに必要なフィールドのみを先に取得

				// native_language（言語フィルタに必要）
				if len(languages) > 0 {
					var nativeLangResults []map[string]interface{}
					err6 := Supabase.DB.From("profiles").
						Select("native_language").
						Eq("user_id", userID).
						Execute(&nativeLangResults)
					// unexpected end of JSON input エラーは無視（プロフィールが存在しない場合）
					if err6 != nil {
						errStr := err6.Error()
						if !containsIgnoreCase(errStr, "unexpected end of json") &&
							!containsIgnoreCase(errStr, "invalid character") {
							log.Printf("SearchUsersAdvanced: error getting native_language for user %s: %v", userID, err6)
						}
					} else if len(nativeLangResults) > 0 {
						profileData["native_language"], _ = nativeLangResults[0]["native_language"].(string)
					}

					// spoken_languages（言語フィルタに必要）
					var spokenLangResults []map[string]interface{}
					err7 := Supabase.DB.From("profiles").
						Select("spoken_languages").
						Eq("user_id", userID).
						Execute(&spokenLangResults)
					if err7 != nil {
						errStr := err7.Error()
						if !containsIgnoreCase(errStr, "unexpected end of json") &&
							!containsIgnoreCase(errStr, "invalid character") {
							log.Printf("SearchUsersAdvanced: error getting spoken_languages for user %s: %v", userID, err7)
						}
					} else if len(spokenLangResults) > 0 {
						profileData["spoken_languages"] = spokenLangResults[0]["spoken_languages"]
					}

					// learning_languages（言語フィルタに必要）
					var learningLangResults []map[string]interface{}
					err8 := Supabase.DB.From("profiles").
						Select("learning_languages").
						Eq("user_id", userID).
						Execute(&learningLangResults)
					if err8 != nil {
						errStr := err8.Error()
						if !containsIgnoreCase(errStr, "unexpected end of json") &&
							!containsIgnoreCase(errStr, "invalid character") {
							log.Printf("SearchUsersAdvanced: error getting learning_languages for user %s: %v", userID, err8)
						}
					} else if len(learningLangResults) > 0 {
						profileData["learning_languages"] = learningLangResults[0]["learning_languages"]
					}
				}

				// residence（国フィルタに必要、または言語フィルタのみの場合でもレスポンスに含めるため取得）
				// フロントエンドで国旗表示のために必要
				var residenceResults []map[string]interface{}
				err4 := Supabase.DB.From("profiles").
					Select("residence").
					Eq("user_id", userID).
					Execute(&residenceResults)
				if err4 != nil {
					errStr := err4.Error()
					if !containsIgnoreCase(errStr, "unexpected end of json") &&
						!containsIgnoreCase(errStr, "invalid character") {
						log.Printf("SearchUsersAdvanced: error getting residence for user %s: %v", userID, err4)
					} else {
						// residenceが存在しない場合の警告ログ（デバッグ用）
						log.Printf("SearchUsersAdvanced: warning - residence not found for user %s", userID)
					}
				} else if len(residenceResults) > 0 {
					profileData["residence"], _ = residenceResults[0]["residence"].(string)
				} else {
					// residenceが存在しない場合の警告ログ（デバッグ用）
					log.Printf("SearchUsersAdvanced: warning - residence not found for user %s", userID)
				}
			}

			// 言語フィルタ
			if len(languages) > 0 {
				matched := false
				nativeLang, _ := profileData["native_language"].(string)
				spokenLangsRaw := profileData["spoken_languages"]
				learningLangsRaw := profileData["learning_languages"]

				// 配列の型変換を安全に行う
				var spokenLangs []interface{}
				if spokenLangsRaw != nil {
					if arr, ok := spokenLangsRaw.([]interface{}); ok {
						spokenLangs = arr
					} else if str, ok := spokenLangsRaw.(string); ok {
						// JSON文字列の場合、パースを試みる
						var parsed []interface{}
						if err := json.Unmarshal([]byte(str), &parsed); err == nil {
							spokenLangs = parsed
						}
					}
				}

				var learningLangs []interface{}
				if learningLangsRaw != nil {
					if arr, ok := learningLangsRaw.([]interface{}); ok {
						learningLangs = arr
					} else if str, ok := learningLangsRaw.(string); ok {
						// JSON文字列の場合、パースを試みる
						var parsed []interface{}
						if err := json.Unmarshal([]byte(str), &parsed); err == nil {
							learningLangs = parsed
						}
					}
				}

				for _, lang := range languages {
					// native_language, spoken_languages, learning_languages のいずれかに一致
					if containsIgnoreCase(nativeLang, lang) {
						matched = true
						break
					}
					for _, spoken := range spokenLangs {
						if spokenStr, ok := spoken.(string); ok && containsIgnoreCase(spokenStr, lang) {
							matched = true
							break
						}
					}
					if matched {
						break
					}
					for _, learning := range learningLangs {
						if learningStr, ok := learning.(string); ok && containsIgnoreCase(learningStr, lang) {
							matched = true
							break
						}
					}
					if matched {
						break
					}
				}
				if !matched {
					return
				}
			}

			// 国フィルタ（residence）
			// 注意: フロントエンドからは英語の国コード（ISO 3166-1 alpha-2）が送られてくる
			// 例: "CN"（中国）, "US"（アメリカ）, "JP"（日本）
			// データベースのresidenceフィールドも同じ形式で保存されていることを想定
			if len(countries) > 0 {
				matched := false
				residence, _ := profileData["residence"].(string)
				if residence == "" {
					// residenceが取得できなかった場合は、このユーザーはマッチしない
					return
				}
				for _, country := range countries {
					// 大文字小文字を区別せずに比較（国コードは通常大文字だが、念のため）
					if strings.EqualFold(residence, country) {
						matched = true
						break
					}
				}
				if !matched {
					return
				}
			}

			// 結果を構築
			result := SearchUserResult{
				UserID:   userID,
				Username: username,
			}

			// プロフィール情報を設定
			// フィルター条件がある場合: フィルタリングに使用したプロフィール情報を設定
			// フィルター条件がない場合: 全ユーザーを返すため、プロフィール情報も取得
			if hasFilters {
				// フィルター条件がある場合: フィルタリングに使用したプロフィール情報を設定
				// residenceは言語フィルタのみの場合でもレスポンスに含める（フロントエンドで国旗表示のため）
				if residence, ok := profileData["residence"].(string); ok && residence != "" {
					result.Residence = residence
				}

				// commentとavatar_urlは結果に含めるが、フィルタリングには不要なので後で取得
				var commentResults []map[string]interface{}
				err11 := Supabase.DB.From("profiles").
					Select("comment").
					Eq("user_id", userID).
					Execute(&commentResults)
				if err11 != nil {
					errStr := err11.Error()
					if !containsIgnoreCase(errStr, "unexpected end of json") &&
						!containsIgnoreCase(errStr, "invalid character") {
						log.Printf("SearchUsersAdvanced: error getting comment for user %s: %v", userID, err11)
					}
				} else if len(commentResults) > 0 {
					result.Comment, _ = commentResults[0]["comment"].(string)
				}

				var avatarResults []map[string]interface{}
				err12 := Supabase.DB.From("profiles").
					Select("avatar_url").
					Eq("user_id", userID).
					Execute(&avatarResults)
				if err12 != nil {
					errStr := err12.Error()
					if !containsIgnoreCase(errStr, "unexpected end of json") &&
						!containsIgnoreCase(errStr, "invalid character") {
						log.Printf("SearchUsersAdvanced: error getting avatar_url for user %s: %v", userID, err12)
					}
				} else if len(avatarResults) > 0 {
					if avatarURL, ok := avatarResults[0]["avatar_url"].(string); ok && avatarURL != "" {
						result.AvatarURL = &avatarURL
					}
					// avatar_urlがNULLまたは空文字列の場合は、result.AvatarURLはnilのまま（JSONではnullとして返される）
				}
			} else {
				// フィルター条件がない場合: 全ユーザーを返すため、プロフィール情報も取得
				// ただし、パフォーマンスを考慮し、必要最小限の情報のみ取得
				var commentResults []map[string]interface{}
				err11 := Supabase.DB.From("profiles").
					Select("comment").
					Eq("user_id", userID).
					Execute(&commentResults)
				if err11 != nil {
					errStr := err11.Error()
					if !containsIgnoreCase(errStr, "unexpected end of json") &&
						!containsIgnoreCase(errStr, "invalid character") {
						log.Printf("SearchUsersAdvanced: error getting comment for user %s: %v", userID, err11)
					}
				} else if len(commentResults) > 0 {
					result.Comment, _ = commentResults[0]["comment"].(string)
				}

				var residenceResults []map[string]interface{}
				err12 := Supabase.DB.From("profiles").
					Select("residence").
					Eq("user_id", userID).
					Execute(&residenceResults)
				if err12 != nil {
					errStr := err12.Error()
					if !containsIgnoreCase(errStr, "unexpected end of json") &&
						!containsIgnoreCase(errStr, "invalid character") {
						log.Printf("SearchUsersAdvanced: error getting residence for user %s: %v", userID, err12)
					}
				} else if len(residenceResults) > 0 {
					result.Residence, _ = residenceResults[0]["residence"].(string)
				}

				var avatarResults []map[string]interface{}
				err13 := Supabase.DB.From("profiles").
					Select("avatar_url").
					Eq("user_id", userID).
					Execute(&avatarResults)
				if err13 != nil {
					errStr := err13.Error()
					if !containsIgnoreCase(errStr, "unexpected end of json") &&
						!containsIgnoreCase(errStr, "invalid character") {
						log.Printf("SearchUsersAdvanced: error getting avatar_url for user %s: %v", userID, err13)
					}
				} else if len(avatarResults) > 0 {
					if avatarURL, ok := avatarResults[0]["avatar_url"].(string); ok && avatarURL != "" {
						result.AvatarURL = &avatarURL
					}
					// avatar_urlがNULLまたは空文字列の場合は、result.AvatarURLはnilのまま（JSONではnullとして返される）
				}
			}

			// 興味情報を取得（パフォーマンス最適化: フィルター条件がない場合は省略）
			// フィルター条件がない場合、興味情報の取得は省略してパフォーマンスを優先
			if hasFilters {
				var userInterestIDs []map[string]interface{}
				err9 := Supabase.DB.From("user_interests").
					Select("interest_id").
					Eq("user_id", userID).
					Execute(&userInterestIDs)
				if err9 == nil && len(userInterestIDs) > 0 {
					// 興味IDのリストを構築
					interestIDs := make([]int, 0, len(userInterestIDs))
					for _, uiMap := range userInterestIDs {
						if interestIDFloat, ok := uiMap["interest_id"].(float64); ok {
							interestID := int(interestIDFloat)
							if interestID > 0 {
								interestIDs = append(interestIDs, interestID)
							}
						}
					}

					// 興味情報を取得（各IDごとに個別クエリ - 最適化の余地あり）
					// ただし、パフォーマンスを考慮し、最大3件までに制限
					maxInterests := 3
					for i, interestID := range interestIDs {
						if i >= maxInterests {
							break
						}
						var interestInfo []map[string]interface{}
						err10 := Supabase.DB.From("interests").
							Select("id, name").
							Eq("id", strconv.Itoa(interestID)).
							Execute(&interestInfo)
						if err10 == nil && len(interestInfo) > 0 {
							if interestName, ok := interestInfo[0]["name"].(string); ok && interestName != "" {
								result.Interests = append(result.Interests, InterestItem{
									ID:   interestID,
									Name: interestName,
								})
							}
						}
					}
				}
			}

			// スレッドセーフに結果を追加
			mu.Lock()
			searchResults = append(searchResults, result)
			mu.Unlock()
		}(userIDMap)
	}

	// すべてのgoroutineの完了を待つ
	wg.Wait()

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
