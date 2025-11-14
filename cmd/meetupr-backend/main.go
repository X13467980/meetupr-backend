package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"meetupr-backend/internal/auth"
	"meetupr-backend/internal/db"
	"meetupr-backend/internal/handlers"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize the authentication service
	auth.Init()

	// Initialize database connection
	db.InitDB()
	defer db.CloseDB() // Close DB connection when main exits

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
	wsGroup := e.Group("/ws")
	wsGroup.Use(auth.EchoJWTMiddleware())
	wsGroup.GET("/chat/:chatID", func(c echo.Context) error {
		// Extract chatID from path parameter
		chatIDStr := c.Param("chatID")
		chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid chat ID")
		}

		// Extract userID from JWT context (set by auth.EchoJWTMiddleware)
		userID, ok := c.Get("user_id").(string)
		if !ok || userID == "" {
			return c.String(http.StatusUnauthorized, "Unauthorized")
		}

		// Pass hub, chatID, and userID to WsHandler
		handlers.WsHandler(hub, chatID, userID, c.Response(), c.Request())
		return nil
	})

	log.Println("Server starting on port 8080...")
	if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}