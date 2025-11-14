package auth

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
)

var (
	jwks     *keyfunc.JWKS
	audience string
	issuer   string
)

func Init() {
	auth0Domain := os.Getenv("AUTH0_DOMAIN")
	if auth0Domain == "" {
		log.Fatal("AUTH0_DOMAIN environment variable not set")
	}
	audience = os.Getenv("AUTH0_AUDIENCE")
	if audience == "" {
		log.Fatal("AUTH0_AUDIENCE environment variable not set")
	}

	issuer = "https://" + auth0Domain + "/"
	jwksURL := "https://" + auth0Domain + "/.well-known/jwks.json"

	var err error
	options := keyfunc.Options{
		Ctx: context.Background(),
		RefreshErrorHandler: func(err error) {
			log.Printf("There was an error with the jwt.Keyfunc\nError: %s", err.Error())
		},
		RefreshInterval:   time.Hour,
		RefreshRateLimit:  time.Minute * 5,
		RefreshTimeout:    time.Second * 10,
		RefreshUnknownKID: true,
	}
	jwks, err = keyfunc.Get(jwksURL, options)
	if err != nil {
		log.Fatalf("Failed to create JWKS from resource at the given URL.\nError: %s", err.Error())
	}
}

// EchoJWTMiddleware creates an Echo middleware for JWT authentication.
func EchoJWTMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Development mode: bypass authentication if DISABLE_AUTH is set
			if os.Getenv("DISABLE_AUTH") == "true" {
				log.Println("⚠️  WARNING: Authentication is DISABLED (development mode)")
				
				// Get user ID from X-Test-User-ID header or use default
				testUserID := c.Request().Header.Get("X-Test-User-ID")
				if testUserID == "" {
					testUserID = "auth0|6917784d99703fe24aebd01d" // Default test user
				}
				
				// Get email from X-Test-User-Email header or use default
				testUserEmail := c.Request().Header.Get("X-Test-User-Email")
				if testUserEmail == "" {
					testUserEmail = "testuser1@example.com"
				}
				
				c.Set("user_id", testUserID)
				c.Set("user_email", testUserEmail)
				return next(c)
			}

			// Production mode: normal JWT authentication
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Authorization header required")
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				return echo.NewHTTPError(http.StatusUnauthorized, "Could not find bearer token in Authorization header")
			}

			// Define custom claims to extract email
			claims := jwt.MapClaims{}

			// Parse and validate the token
			token, err := jwt.ParseWithClaims(tokenString, claims, jwks.Keyfunc)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Failed to parse or validate token: "+err.Error())
			}

			if !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
			}

			// --- Set user info into context ---
			userID, ok := claims["sub"].(string)
			if !ok || userID == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "User ID (sub) not found in token claims")
			}
			c.Set("user_id", userID)

			// Extract email (assuming it's in a custom claim, adjust if necessary)
			// The claim name depends on your Auth0 configuration (Rules or Actions)
			// It might be "email", "https://example.com/email", etc.
			userEmail, ok := claims["https://meetupr.com/email"].(string)
			if !ok {
				// Fallback for the standard email claim, which might not be present
				// depending on the OIDC configuration and requested scopes.
				userEmail, _ = claims["email"].(string)
			}
			c.Set("user_email", userEmail)


			return next(c)
		}
	}
}