package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

	// Initialize Echo
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Initialize and run the ChatHub
	hub := handlers.NewHub()
	go hub.Run()

	// Public routes
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	// WebSocket route with JWT middleware
	e.GET("/ws", func(c echo.Context) error {
		handlers.WsHandler(hub, c.Response(), c.Request())
		return nil
	}, auth.EchoJWTMiddleware())

	log.Println("Server starting on port 8080...")
	if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}