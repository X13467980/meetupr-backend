package handlers

import (
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
