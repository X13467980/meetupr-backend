package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"meetupr-backend/internal/db"

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
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
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

			if roomClients, ok := h.rooms[msg.ChatID]; ok {
				for client := range roomClients {
					select {
					case client.send <- jsonMessage:
					default:
						close(client.send)
						delete(h.clients, client)
						delete(roomClients, client)
					}
				}
			}
		}
	}
}

// WsHandler handles websocket requests from the peer.
func WsHandler(hub *Hub, w http.ResponseWriter, r *http.Request, chatID int64, userID string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		chatID: chatID,
		userID: userID,
	}
	client.hub.register <- client

	// Load and send message history
	go loadMessageHistory(client)

	go client.writePump()
	go client.readPump()
}

func saveMessage(m *Message) error {
	query := `INSERT INTO messages (chat_id, sender_id, content, message_type) VALUES ($1, $2, $3, $4)`
	_, err := db.DB.Exec(query, m.ChatID, m.SenderID, m.Content, "text") // Assuming message_type is always "text" for now
	return err
}

func loadMessageHistory(c *Client) {
	query := `SELECT content, sender_id FROM messages WHERE chat_id = $1 ORDER BY sent_at ASC`

rows, err := db.DB.Query(query, c.chatID)
	if err != nil {
		log.Printf("error loading message history: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var content, senderID string
		if err := rows.Scan(&content, &senderID); err != nil {
			log.Printf("error scanning message row: %v", err)
			continue
		}
		msg := Message{
			Content:  content,
			ChatID:   c.chatID,
			SenderID: senderID,
		}
		jsonMessage, err := json.Marshal(msg)
		if err != nil {
			log.Printf("error marshalling history message: %v", err)
			continue
		}
		c.send <- jsonMessage
	}
}