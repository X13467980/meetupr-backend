package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestWebSocketEcho(t *testing.T) {
	// Create a new Hub for testing
	hub := NewHub()
	go hub.Run()

	// Test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WsHandler(hub, w, r)
	}))
	defer server.Close()

	// Convert http:// to ws://
	u := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect to the server
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer ws.Close()

	// Send test message
	testMessage := "hello"
	if err := ws.WriteMessage(websocket.TextMessage, []byte(testMessage)); err != nil {
		t.Fatalf("WriteMessage failed: %v", err)
	}

	// Receive echo message (from broadcast)
	_, p, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage failed: %v", err)
	}

	// Check if the message is correct
	if string(p) != testMessage {
		t.Errorf("handler returned unexpected body: got %v want %v", string(p), testMessage)
	}

	// Send close message
	err = ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		t.Fatalf("Write close message failed: %v", err)
	}
}
