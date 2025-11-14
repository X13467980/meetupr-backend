package handlers

import (
	"net/http"
	"strconv"

	"meetupr-backend/internal/db"

	"github.com/labstack/echo/v4"
)

// GetChats godoc
// @Summary Get user's chat rooms
// @Description 自身が参加しているチャットルームの一覧を取得します。
// @Tags chats
// @Produce  json
// @Success 200 {array} map[string]interface{}
// @Failure 401 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /api/v1/chats [get]
// GetChats: JWT から user_id を取得し、db.GetChats を呼んで参加チャットルームの一覧を返します。
func GetChats(c echo.Context) error {
	// 認証済みの user_id を取得
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in token")
	}

	// チャットルーム一覧を取得
	chats, err := db.GetChats(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get chats: "+err.Error())
	}

	return c.JSON(http.StatusOK, chats)
}

// GetChatMessages godoc
// @Summary Get chat messages
// @Description 特定のチャットルームのメッセージ履歴を取得します。
// @Tags chats
// @Produce  json
// @Param   chatId path int64 true "Chat ID"
// @Success 200 {array} map[string]interface{}
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /api/v1/chats/{chatId}/messages [get]
// GetChatMessages: パスパラメータから chatId を取得し、db.GetChatMessages を呼んでメッセージ履歴を返します。
func GetChatMessages(c echo.Context) error {
	// 認証確認
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in token")
	}

	// chatId をパスパラメータから取得
	chatIDStr := c.Param("chatId")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid chat ID")
	}

	// メッセージ履歴を取得
	messages, err := db.GetChatMessages(chatID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get messages: "+err.Error())
	}

	return c.JSON(http.StatusOK, messages)
}
