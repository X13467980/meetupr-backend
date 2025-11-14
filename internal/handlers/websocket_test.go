package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"meetupr-backend/internal/db"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

// Helper function to create a test server and a client connection
func newTestClient(t *testing.T, hub *Hub, chatID int64, userID string) (*websocket.Conn, func()) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WsHandler(hub, w, r, chatID, userID)
	}))

	u := "ws" + strings.TrimPrefix(server.URL, "http")

	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}

	cleanup := func() {
		ws.Close()
		server.Close()
	}

	return ws, cleanup
}

func TestChatRoomBroadcast(t *testing.T) {
	// Load .env file for database connection
	if err := godotenv.Load("../../.env"); err != nil {
		t.Fatalf("Error loading .env file for tests: %v", err)
	}
	db.Init()

	hub := NewHub()
	go hub.Run()

	chatID := int64(1)
	user1ID := "user1"
	user2ID := "user2"

	// Create client 1
	ws1, cleanup1 := newTestClient(t, hub, chatID, user1ID)
	defer cleanup1()

	// Create client 2
	ws2, cleanup2 := newTestClient(t, hub, chatID, user2ID)
	defer cleanup2()

	// Give the hub a moment to register the clients
	time.Sleep(100 * time.Millisecond)

	// Client 1 sends a message
	testContent := "hello from user1"
	if err := ws1.WriteMessage(websocket.TextMessage, []byte(testContent)); err != nil {
		t.Fatalf("Client 1 WriteMessage failed: %v", err)
	}

	// Client 2 should receive the message
	_, p, err := ws2.ReadMessage()
	if err != nil {
		t.Fatalf("Client 2 ReadMessage failed: %v", err)
	}

	var msg Message
	if err := json.Unmarshal(p, &msg); err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}

	if msg.Content != testContent {
		t.Errorf("Client 2 received wrong content: got %v want %v", msg.Content, testContent)
	}
	if msg.SenderID != user1ID {
		t.Errorf("Client 2 received wrong sender ID: got %v want %v", msg.SenderID, user1ID)
	}
	if msg.ChatID != chatID {
		t.Errorf("Client 2 received wrong chat ID: got %v want %v", msg.ChatID, chatID)
	}

	// Client 1 should NOT receive its own message back in this simple setup
	// (though some chat apps do this). We'll just check that it doesn't block.
	ws1.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	_, _, err = ws1.ReadMessage()
	if err == nil {
		t.Errorf("Client 1 should not have received a message, but it did")
	}
}
