package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

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

	// CORS middleware
	// 環境変数 CORS_ALLOW_ORIGINS から許可するオリジンを取得（カンマ区切り）
	// 未設定の場合は "*" を許可（開発環境用）
	corsAllowOrigins := os.Getenv("CORS_ALLOW_ORIGINS")
	var allowOrigins []string
	if corsAllowOrigins == "" {
		allowOrigins = []string{"*"}
		log.Println("CORS: Allowing all origins (development mode)")
	} else {
		allowOrigins = strings.Split(corsAllowOrigins, ",")
		log.Printf("CORS: Allowing origins: %v", allowOrigins)
	}

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     allowOrigins,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
	}))

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
	// More specific routes must be defined before the generic one
	chatGroup.GET("/with/:otherUserId", handlers.GetOrCreateChatWithUser, auth.EchoJWTMiddleware())
	chatGroup.GET("/:chatId/messages", handlers.GetChatMessages, auth.EchoJWTMiddleware())
	chatGroup.GET("/:chatId", handlers.GetChatDetail, auth.EchoJWTMiddleware())

	// Search routes
	searchGroup := apiV1.Group("/search")
	searchGroup.GET("/users", handlers.SearchUsersWithQuery, auth.EchoJWTMiddleware())
	searchGroup.POST("/users", handlers.SearchUsersAdvanced, auth.EchoJWTMiddleware())

	// WebSocket route with JWT middleware
	// Note: WebSocket connections typically pass token as query parameter (?token=...)
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

		// Verify that the user is a participant in this chat
		isParticipant, err := db.IsChatParticipant(chatID, userID)
		if err != nil {
			log.Printf("Error verifying chat access: %v", err)
			return c.String(http.StatusInternalServerError, "Failed to verify chat access")
		}
		if !isParticipant {
			return c.String(http.StatusForbidden, "You are not a participant in this chat")
		}

		handlers.WsHandler(hub, c.Response(), c.Request(), chatID, userID)
		return nil
	}, auth.EchoJWTMiddleware())

	// Get port from environment variable (Render sets PORT env var)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port for local development
	}

	log.Printf("Server starting on port %s...", port)
	if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
