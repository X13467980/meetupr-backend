package handlers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"meetupr-backend/internal/db"
	"meetupr-backend/internal/models"
	"strconv"
)

// RegisterUser godoc
// @Summary Register a new user
// @Description Register a new user after Auth0 authentication
// @Tags users
// @Accept  json
// @Produce  json
// @Param   user body models.RegisterUserRequest true "User registration info"
// @Success 201 {object} models.User
// @Failure 409 {object} echo.HTTPError
// @Router /api/v1/users/register [post]
func RegisterUser(c echo.Context) error {
	var req models.RegisterUserRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in token")
	}

	userEmail, ok := c.Get("user_email").(string)
	if !ok || userEmail == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "User email not found in token")
	}

	user := models.User{
		ID:            userID,
		Email:         userEmail,
		Username:      req.Username,
		IsOICVerified: false,
		CreatedAt:     time.Now(),
	}

	if err := db.CreateUser(user); err != nil {
		// Check for unique constraint violation (e.g., username or email already exists)
		// This is a simplified check. In a real app, you'd parse the specific DB error.
		if err.Error() == "PGRST202: duplicate key value violates unique constraint \"users_email_key\"" ||
			err.Error() == "PGRST202: duplicate key value violates unique constraint \"users_username_key\"" {
			return echo.NewHTTPError(http.StatusConflict, "User with this email or username already exists")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to register user: "+err.Error())
	}

	return c.JSON(http.StatusCreated, user)
}

// GetMyProfile godoc
// @Summary Get current user's profile
// @Description Get the detailed profile of the currently logged-in user
// @Tags users
// @Produce  json
// @Success 200 {object} models.UserProfileResponse
// @Router /api/v1/users/me [get]
func GetMyProfile(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in token")
	}

	profile, err := db.GetUserByID(userID)
	if err != nil {
		// Check if the error is a "not found" error from Supabase
		if err.Error() == "PGRST116: The record you are attempting to retrieve was not found." {
			return echo.NewHTTPError(http.StatusNotFound, "User profile not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get user profile: "+err.Error())
	}

	return c.JSON(http.StatusOK, profile)
}

// UpdateMyProfile godoc
// @Summary Update current user's profile
// @Description Update the profile of the currently logged-in user
// @Tags users
// @Accept  json
// @Produce  json
// @Param   profile body models.UpdateUserProfileRequest true "User profile update info"
// @Success 200 {object} models.UserProfileResponse
// @Router /api/v1/users/me [put]
func UpdateMyProfile(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in token")
	}

	var req models.UpdateUserProfileRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	updatedProfile, err := db.UpdateUserProfile(userID, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update user profile: "+err.Error())
	}

	return c.JSON(http.StatusOK, updatedProfile)
}

// SearchUsers godoc
// @Summary Search for users
// @Description Search for other users based on criteria
// @Tags users
// @Produce  json
// @Param   interest_id query int false "Interest ID"
// @Param   learning_language query string false "Learning Language"
// @Param   spoken_language query string false "Spoken Language"
// @Success 200 {array} models.User
// @Router /api/v1/users [get]
func SearchUsers(c echo.Context) error {
	interestIDStr := c.QueryParam("interest_id")
	learningLanguage := c.QueryParam("learning_language")
	spokenLanguage := c.QueryParam("spoken_language")

	var interestID int
	if interestIDStr != "" {
		parsedID, err := strconv.Atoi(interestIDStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid interest_id")
		}
		interestID = parsedID
	}

	users, err := db.SearchUsers(interestID, learningLanguage, spokenLanguage)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to search users: "+err.Error())
	}

	return c.JSON(http.StatusOK, users)
}

// GetUserProfile godoc
// @Summary Get a user's public profile
// @Description Get the public profile of a specific user
// @Tags users
// @Produce  json
// @Param   userId path string true "User ID"
// @Success 200 {object} models.User
// @Router /api/v1/users/{userId} [get]
func GetUserProfile(c echo.Context) error {
	userID := c.Param("userId")

	user, err := db.GetUserProfile(userID)
	if err != nil {
		// Check if the error is a "not found" error from Supabase
		if err.Error() == "PGRST116: The record you are attempting to retrieve was not found." {
			return echo.NewHTTPError(http.StatusNotFound, "User not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get user profile: "+err.Error())
	}

	return c.JSON(http.StatusOK, user)
}
