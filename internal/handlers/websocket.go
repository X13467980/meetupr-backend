package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // For development, allow all origins
	},
}

// Message represents a message sent over the websocket.
type Message struct {
	Type    string    `json:"type"` // e.g., "text", "join", "leave"
	ChatID  int64     `json:"chat_id"`
	SenderID string    `json:"sender_id"`
	Content string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// User ID of the client
	userID string

	// Chat ID the client is currently in
	chatID int64
}

// readPump pumps messages from the websocket connection to the hub.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, p, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(p, &msg); err != nil {
			log.Printf("error unmarshalling message: %v", err)
			continue
		}

		// Ensure message has correct chat ID and sender ID
		msg.ChatID = c.chatID
		msg.SenderID = c.userID
		msg.Timestamp = time.Now()

		c.hub.broadcast <- msg // broadcast channel now takes Message struct
	}
}

// writePump pumps messages from the hub to the websocket connection.
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
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte("\n"))
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
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan Message // Changed to Message struct

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// Rooms map: chatID -> map of clients in that room
	rooms map[int64]map[*Client]bool
}

// NewHub creates and returns a new Hub instance.
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan Message), // Changed to Message struct
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		rooms:      make(map[int64]map[*Client]bool),
	}
}

// Run starts the hub's event loop.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			if _, ok := h.rooms[client.chatID]; !ok {
				h.rooms[client.chatID] = make(map[*Client]bool)
			}
			h.rooms[client.chatID][client] = true
			log.Printf("Client %s joined chat %d", client.userID, client.chatID)

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				if roomClients, ok := h.rooms[client.chatID]; ok {
					delete(roomClients, client)
					if len(roomClients) == 0 {
						delete(h.rooms, client.chatID) // Remove room if empty
					}
				}
				log.Printf("Client %s left chat %d", client.userID, client.chatID)
			}

		case message := <-h.broadcast:
			// Marshal message back to JSON for sending
			jsonMessage, err := json.Marshal(message)
			if err != nil {
				log.Printf("error marshalling message: %v", err)
				continue
			}

			if roomClients, ok := h.rooms[message.ChatID]; ok {
				for client := range roomClients {
					select {
					case client.send <- jsonMessage:
					default:
						close(client.send)
						delete(h.clients, client)
						delete(roomClients, client)
						if len(roomClients) == 0 {
							delete(h.rooms, message.ChatID)
						}
					}
				}
			}
		}
	}
}

// WsHandler handles WebSocket connections.
// It now expects chatID and userID to be extracted from the request context or path.
func WsHandler(hub *Hub, chatID int64, userID string, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256), userID: userID, chatID: chatID}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}
