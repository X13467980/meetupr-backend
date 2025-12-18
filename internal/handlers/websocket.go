package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"meetupr-backend/internal/db"
	"meetupr-backend/internal/models"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Message defines the structure for messages sent over WebSocket.
type Message struct {
	Content  string `json:"content"`
	ChatID   int64  `json:"chat_id"`
	SenderID string `json:"sender_id"`
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	chatID int64
	userID string
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		// Read message from browser
		_, rawMessage, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// Unmarshal the raw message to extract content
		var msgData map[string]string
		if err := json.Unmarshal(rawMessage, &msgData); err != nil {
			log.Printf("error unmarshalling raw message: %v", err)
			continue
		}

		// Create a message struct and populate it
		msg := Message{
			Content:  msgData["content"],
			ChatID:   c.chatID,
			SenderID: c.userID,
		}

		// Marshal the struct to JSON to be sent to the hub
		jsonMessage, err := json.Marshal(msg)
		if err != nil {
			log.Printf("error marshalling message: %v", err)
			continue
		}

		c.hub.broadcast <- jsonMessage
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
		log.Printf("writePump: connection closed for chat %d, user %s", c.chatID, c.userID)
	}()

	messageCount := 0
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				log.Printf("writePump: send channel closed for chat %d, user %s", c.chatID, c.userID)
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Write message as a separate WebSocket message (not batched with newlines)
			// This makes it easier for frontend to parse each message individually
			err := c.conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Printf("writePump: failed to write message for chat %d: %v", c.chatID, err)
				return
			}
			messageCount++

			// Try to unmarshal to log content
			var msgPreview models.Message
			if err := json.Unmarshal(message, &msgPreview); err == nil {
				log.Printf("writePump: wrote message %d to WebSocket for chat %d: id=%d, content=%s", messageCount, c.chatID, msgPreview.ID, msgPreview.Content)
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("writePump: failed to send ping for chat %d: %v", c.chatID, err)
				return
			}
		}
	}
}

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	// Maps chatID to a set of clients in that room.
	rooms map[int64]map[*Client]bool
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		rooms:      make(map[int64]map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			if h.rooms[client.chatID] == nil {
				h.rooms[client.chatID] = make(map[*Client]bool)
			}
			h.rooms[client.chatID][client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				if roomClients := h.rooms[client.chatID]; roomClients != nil {
					delete(roomClients, client)
					if len(roomClients) == 0 {
						delete(h.rooms, client.chatID)
					}
				}
			}
		case jsonMessage := <-h.broadcast:
			var msg Message
			if err := json.Unmarshal(jsonMessage, &msg); err != nil {
				log.Printf("error unmarshalling broadcast message: %v", err)
				continue
			}

			// Save message to the database
			if err := saveMessage(&msg); err != nil {
				log.Printf("error saving message to db: %v", err)
				continue
			}
			log.Printf("Message saved to DB: chat_id=%d, sender_id=%s, content=%s", msg.ChatID, msg.SenderID, msg.Content)

			// Convert handlers.Message to models.Message format for frontend
			fullMsg := models.Message{
				ChatID:      msg.ChatID,
				SenderID:    msg.SenderID,
				Content:     msg.Content,
				MessageType: "text",
				SentAt:      time.Now(),
			}
			messageToBroadcast, err := json.Marshal(fullMsg)
			if err != nil {
				log.Printf("error marshalling message: %v", err)
				continue
			}

			if roomClients, ok := h.rooms[msg.ChatID]; ok {
				log.Printf("Broadcasting message to %d client(s) in chat %d", len(roomClients), msg.ChatID)
				for client := range roomClients {
					select {
					case client.send <- messageToBroadcast:
						log.Printf("Message sent to client: chat_id=%d, user_id=%s", client.chatID, client.userID)
					default:
						log.Printf("Client send channel full, removing client: user_id=%s", client.userID)
						close(client.send)
						delete(h.clients, client)
						delete(roomClients, client)
					}
				}
			} else {
				log.Printf("No clients found in chat room %d", msg.ChatID)
			}
		}
	}
}

// WsHandler handles websocket requests from the peer.
func WsHandler(hub *Hub, w http.ResponseWriter, r *http.Request, chatID int64, userID string) {
	log.Printf("WsHandler: WebSocket connection attempt for chat %d, user %s", chatID, userID)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WsHandler: failed to upgrade connection: %v", err)
		return
	}
	log.Printf("WsHandler: WebSocket connection established for chat %d, user %s", chatID, userID)

	client := &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		chatID: chatID,
		userID: userID,
	}
	client.hub.register <- client
	log.Printf("WsHandler: client registered for chat %d", chatID)

	// Load and send message history
	go loadMessageHistory(client)

	go client.writePump()
	go client.readPump()
}

func saveMessage(m *Message) error {
	// Supabase API経由でメッセージを挿入
	messageData := map[string]interface{}{
		"chat_id":      m.ChatID,
		"sender_id":    m.SenderID,
		"content":      m.Content,
		"message_type": "text", // Assuming message_type is always "text" for now
	}

	var results []map[string]interface{}
	err := db.Supabase.DB.From("messages").Insert(messageData).Execute(&results)
	return err
}

func loadMessageHistory(c *Client) {
	// Use db.GetChatMessages to get message history
	messages, err := db.GetChatMessages(c.chatID)
	if err != nil {
		log.Printf("error loading message history from Supabase: %v", err)
		return
	}

	log.Printf("loadMessageHistory: Loaded %d message(s) from history for chat %d", len(messages), c.chatID)

	if len(messages) == 0 {
		log.Printf("loadMessageHistory: No messages found for chat %d", c.chatID)
		return
	}

	// Send messages in the format that frontend expects (models.Message format)
	sentCount := 0
	for _, dbMsg := range messages {
		// Use the full models.Message structure with all fields
		jsonMessage, err := json.Marshal(dbMsg)
		if err != nil {
			log.Printf("loadMessageHistory: error marshalling history message: %v", err)
			continue
		}
		select {
		case c.send <- jsonMessage:
			sentCount++
			log.Printf("loadMessageHistory: Sent history message %d/%d: id=%d, content=%s", sentCount, len(messages), dbMsg.ID, dbMsg.Content)
		case <-time.After(5 * time.Second):
			log.Printf("loadMessageHistory: timeout sending message %d, client send channel may be blocked", dbMsg.ID)
			break
		default:
			log.Printf("loadMessageHistory: client send channel full, skipping history message %d", dbMsg.ID)
		}
	}
	log.Printf("loadMessageHistory: Finished sending %d/%d message(s) from history for chat %d", sentCount, len(messages), c.chatID)
}
