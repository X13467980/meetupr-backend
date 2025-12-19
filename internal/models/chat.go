package models

import "time"

// Chat represents a chat room between two users
type Chat struct {
	ID               int64     `json:"id"`
	User1ID          string    `json:"user1_id"`
	User2ID          string    `json:"user2_id"`
	AISuggestedTheme string    `json:"ai_suggested_theme,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	// Additional fields for response
	OtherUser   *User    `json:"other_user,omitempty"`
	LastMessage *Message `json:"last_message,omitempty"`
}

// Message represents a message in a chat
type Message struct {
	ID                int64     `json:"id"`
	ChatID            int64     `json:"chat_id"`
	SenderID          string    `json:"sender_id"`
	Content           string    `json:"content"`
	TranslatedContent string    `json:"translated_content,omitempty"`
	MessageType       string    `json:"message_type"`
	SentAt            time.Time `json:"sent_at"`
}
