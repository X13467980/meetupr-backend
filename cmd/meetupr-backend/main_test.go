package main

import (
	"meetupr-backend/internal/handlers"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestWebSocketEcho(t *testing.T) {
	// Create a new Hub for testing
	hub := handlers.NewHub()
	go hub.Run()

	// Test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.WsHandler(hub, w, r)
	}))
	defer server.Close()

	// Convert http:// to ws://
	u := "ws" + strings.TrimPrefix(server.URL, "http")

	// WebSocketサーバーに接続
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer ws.Close()

	// テストメッセージを送信
	testMessage := "hello"
	if err := ws.WriteMessage(websocket.TextMessage, []byte(testMessage)); err != nil {
		t.Fatalf("WriteMessage failed: %v", err)
	}

	// サーバーからのエコーメッセージを受信
	_, p, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage failed: %v", err)
	}

	// 受信したメッセージが送信したメッセージと一致するか確認
	if string(p) != testMessage {
		t.Errorf("handler returned unexpected body: got %v want %v", string(p), testMessage)
	}

	// サーバーに正常なクローズメッセージを送信
	err = ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		t.Fatalf("Write close message failed: %v", err)
	}
}
