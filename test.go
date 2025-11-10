package main

import (
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

func testWebSocketClient() {
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	message := []byte("Hello, WebSocket!")
	err = c.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		log.Println("write:", err)
		return
	}
	log.Printf("sent: %s", message)

	time.Sleep(time.Second)

	err = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("write close:", err)
		return
	}

	select {
	case <-done:
	case <-time.After(time.Second):
	}
	log.Println("WebSocket client finished")
}

func main() {
	// main.goでサーバーを起動した後、この関数を呼び出してクライアントテストを実行できます。
	// testWebSocketClient()
}