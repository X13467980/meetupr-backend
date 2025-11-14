package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"meetupr-backend/internal/auth"
	"meetupr-backend/internal/handlers"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize the authentication service
	auth.Init()

	// Initialize and run the ChatHub
	hub := handlers.NewHub()
	go hub.Run()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	})

	// Protect the /ws endpoint with the JWT middleware
	http.Handle("/ws", auth.JWTMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.WsHandler(hub, w, r)
	})))

	log.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}