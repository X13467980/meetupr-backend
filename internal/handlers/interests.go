package handlers

import (
	"net/http"

	"meetupr-backend/internal/db"

	"github.com/labstack/echo/v4"
)

// GetInterests godoc
// @Summary Get interests master data
// @Description Retrieve the list of available interests/categories
// @Tags interests
// @Produce  json
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/interests [get]
func GetInterests(c echo.Context) error {
	interests, err := db.GetInterests()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get interests: "+err.Error())
	}
	return c.JSON(http.StatusOK, interests)
}
