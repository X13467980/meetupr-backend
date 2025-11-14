package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

var (
	addr   = flag.String("addr", "localhost:8080", "http service address")
	chatID = flag.String("chat", "1", "chat ID")
	token  = flag.String("token", "", "JWT token (required)")
	userID = flag.String("user", "", "User ID for display")
)

func main() {
	flag.Parse()

	if *token == "" {
		log.Fatal("JWT token is required. Use -token flag")
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Build WebSocket URL with JWT token in Authorization header
	u := url.URL{
		Scheme: "ws",
		Host:   *addr,
		Path:   fmt.Sprintf("/ws/chat/%s", *chatID),
	}

	log.Printf("Connecting to %s", u.String())

	// Create request headers with JWT token
	headers := make(http.Header)
	headers.Set("Authorization", fmt.Sprintf("Bearer %s", *token))

	// Connect to WebSocket
	c, resp, err := websocket.DefaultDialer.Dial(u.String(), headers)
	if err != nil {
		if resp != nil {
			log.Printf("HTTP Response Status: %s", resp.Status)
			log.Printf("HTTP Response Headers: %v", resp.Header)
		}
		log.Fatal("dial:", err)
	}
	defer c.Close()

	log.Printf("✅ Successfully connected! HTTP Status: %s", resp.Status)
	log.Printf("📡 Waiting for messages... (Press Ctrl+C to exit)")
	log.Printf("💡 Tip: Open another terminal and run the same command with a different user token to test message broadcasting")

	done := make(chan struct{})

	// Read messages from server
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket closed: %v", err)
				} else {
					log.Printf("Read error: %v", err)
				}
				return
			}

			// Try to parse as JSON for better display
			var msgData map[string]interface{}
			if err := json.Unmarshal(message, &msgData); err == nil {
				log.Printf("📨 Received: %+v", msgData)
			} else {
				log.Printf("📨 Received (raw): %s", message)
			}
		}
	}()

	// Send messages
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	// Send initial test message
	testMsg := map[string]string{
		"content": "Hello from test client!",
	}
	msgBytes, err := json.Marshal(testMsg)
	if err != nil {
		log.Fatal("marshal:", err)
	}
	err = c.WriteMessage(websocket.TextMessage, msgBytes)
	if err != nil {
		log.Printf("❌ Write error: %v", err)
		return
	}
	log.Printf("📤 Sent initial message: %s", testMsg["content"])

	// Keep connection alive and send periodic messages
	messageCount := 0
	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			messageCount++
			testMsg := map[string]string{
				"content": fmt.Sprintf("Test message #%d at %s", messageCount, t.Format(time.RFC3339)),
			}
			msgBytes, err := json.Marshal(testMsg)
			if err != nil {
				log.Println("marshal:", err)
				return
			}
			err = c.WriteMessage(websocket.TextMessage, msgBytes)
			if err != nil {
				log.Printf("❌ Write error: %v", err)
				return
			}
			log.Printf("📤 Sent message #%d: %s", messageCount, testMsg["content"])
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
