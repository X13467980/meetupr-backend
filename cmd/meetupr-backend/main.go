package main

import (
	"log"
	"net/http"
	"strconv"

	"meetupr-backend/internal/auth"
	"meetupr-backend/internal/db"
	"meetupr-backend/internal/handlers"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	echoSwagger "github.com/swaggo/echo-swagger"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, proceeding with environment variables")
	}

	// Initialize the authentication service
	auth.Init()

	// Initialize the database connection
	db.Init()

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

	// Serve swagger.yaml
	e.GET("/swagger.yaml", func(c echo.Context) error {
		return c.File("docs/swagger.yaml")
	})

	// Swagger UI route
	e.GET("/swagger/*", echoSwagger.EchoWrapHandler(echoSwagger.URL("/swagger.yaml")))

	// API V1 routes
	apiV1 := e.Group("/api/v1")

	// User routes
	userGroup := apiV1.Group("/users")
	userGroup.POST("/register", handlers.RegisterUser, auth.EchoJWTMiddleware())
	userGroup.GET("/me", handlers.GetMyProfile, auth.EchoJWTMiddleware())
	userGroup.PUT("/me", handlers.UpdateMyProfile, auth.EchoJWTMiddleware())
	userGroup.GET("", handlers.SearchUsers, auth.EchoJWTMiddleware())
	userGroup.GET("/:userId", handlers.GetUserProfile, auth.EchoJWTMiddleware())

	// Interests routes
	interestGroup := apiV1.Group("/interests")
	interestGroup.GET("", handlers.GetInterests, auth.EchoJWTMiddleware())

	// Chat routes
	chatGroup := apiV1.Group("/chats")
	chatGroup.GET("", handlers.GetChats, auth.EchoJWTMiddleware())
	chatGroup.GET("/:chatId/messages", handlers.GetChatMessages, auth.EchoJWTMiddleware())

	// WebSocket route with JWT middleware
	e.GET("/ws/chat/:chatID", func(c echo.Context) error {
		chatIDStr := c.Param("chatID")
		chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid chat ID")
		}

		userID, ok := c.Get("user_id").(string)
		if !ok || userID == "" {
			return c.String(http.StatusUnauthorized, "User ID not found in token")
		}

		handlers.WsHandler(hub, c.Response(), c.Request(), chatID, userID)
		return nil
	}, auth.EchoJWTMiddleware())

	log.Println("Server starting on port 8080...")
	if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
