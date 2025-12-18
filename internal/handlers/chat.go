package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"meetupr-backend/internal/db"

	"github.com/labstack/echo/v4"
)

// GetChats godoc
// @Summary Get list of chat rooms
// @Description Get a list of chat rooms that the current user is participating in
// @Tags chats
// @Produce  json
// @Success 200 {array} models.Chat
// @Router /api/v1/chats [get]
func GetChats(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in token")
	}

	chats, err := db.GetUserChats(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get chats: "+err.Error())
	}

	return c.JSON(http.StatusOK, chats)
}

// GetChatMessages godoc
// @Summary Get messages from a chat room
// @Description Get message history from a specific chat room
// @Tags chats
// @Produce  json
// @Param   chatId path int true "Chat ID"
// @Success 200 {array} models.Message
// @Router /api/v1/chats/{chatId}/messages [get]
func GetChatMessages(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in token")
	}

	chatIDStr := c.Param("chatId")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid chat ID")
	}

	// Verify that the user is a participant in this chat
	isParticipant, err := db.IsChatParticipant(chatID, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to verify chat access: "+err.Error())
	}
	if !isParticipant {
		return echo.NewHTTPError(http.StatusForbidden, "You are not a participant in this chat")
	}

	messages, err := db.GetChatMessages(chatID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get messages: "+err.Error())
	}

	return c.JSON(http.StatusOK, messages)
}

// GetChatDetail godoc
// @Summary Get chat room details
// @Description Get detailed information about a specific chat room including other user information
// @Tags chats
// @Produce  json
// @Param   chatId path int true "Chat ID"
// @Success 200 {object} models.Chat
// @Router /api/v1/chats/{chatId} [get]
func GetChatDetail(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in token")
	}

	chatIDStr := c.Param("chatId")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid chat ID")
	}

	chat, err := db.GetChatDetail(chatID, userID)
	if err != nil {
		if err.Error() == fmt.Sprintf("user %s is not a participant in chat %d", userID, chatID) {
			return echo.NewHTTPError(http.StatusForbidden, "You are not a participant in this chat")
		}
		if err.Error() == fmt.Sprintf("chat %d not found", chatID) {
			return echo.NewHTTPError(http.StatusNotFound, "Chat not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get chat details: "+err.Error())
	}

	return c.JSON(http.StatusOK, chat)
}

// GetOrCreateChatWithUser godoc
// @Summary Get or create a chat with another user
// @Description Get an existing chat ID or create a new chat between the current user and another user
// @Tags chats
// @Produce  json
// @Param   otherUserId path string true "Other User ID"
// @Success 200 {object} map[string]interface{} "Returns chat_id"
// @Router /api/v1/chats/with/{otherUserId} [get]
func GetOrCreateChatWithUser(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in token")
	}

	otherUserID := c.Param("otherUserId")
	if otherUserID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Other user ID is required")
	}

	if otherUserID == userID {
		return echo.NewHTTPError(http.StatusBadRequest, "Cannot create a chat with yourself")
	}

	chatID, err := db.GetOrCreateChat(userID, otherUserID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get or create chat: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"chat_id": chatID,
	})
}
