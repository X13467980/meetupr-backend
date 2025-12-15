package handlers

import (
	"net/http"
	"strings"

	"meetupr-backend/internal/db"

	"github.com/labstack/echo/v4"
)

// SearchRequest represents the search request body
type SearchRequest struct {
	Keyword   string   `json:"keyword"`   // キーワード検索（ユーザー名など）
	Languages []string `json:"languages"` // 言語フィルタ（日本語、英語など）
	Countries []string `json:"countries"` // 国フィルタ（residence）
}

// SearchUserResult represents a single user in search results
type SearchUserResult struct {
	UserID    string         `json:"user_id"`
	Username  string         `json:"username"`
	Comment   string         `json:"comment"`
	Residence string         `json:"residence"`
	Interests []InterestItem `json:"interests"`
}

// InterestItem represents an interest/hobby item
type InterestItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// SearchUsersAdvanced godoc
// @Summary Advanced search for users
// @Description Search for users with keyword, language, and country filters
// @Tags search
// @Accept  json
// @Produce  json
// @Param   search body SearchRequest true "Search parameters"
// @Success 200 {array} SearchUserResult
// @Router /api/v1/search/users [post]
func SearchUsersAdvanced(c echo.Context) error {
	// 自身のユーザーIDを取得
	currentUserID, ok := c.Get("user_id").(string)
	if !ok || currentUserID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in token")
	}

	var req SearchRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// 検索パラメータの正規化（空白トリム）
	req.Keyword = strings.TrimSpace(req.Keyword)
	for i := range req.Languages {
		req.Languages[i] = strings.TrimSpace(req.Languages[i])
	}
	for i := range req.Countries {
		req.Countries[i] = strings.TrimSpace(req.Countries[i])
	}

	// DB検索実行
	results, err := db.SearchUsersAdvanced(currentUserID, req.Keyword, req.Languages, req.Countries)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to search users: "+err.Error())
	}

	return c.JSON(http.StatusOK, results)
}

// SearchUsersWithQuery godoc
// @Summary Search for users with query parameters
// @Description Search for users using GET request with query parameters
// @Tags search
// @Produce  json
// @Param   keyword query string false "Keyword to search in username"
// @Param   language query string false "Language filter (comma-separated)"
// @Param   country query string false "Country/residence filter (comma-separated)"
// @Success 200 {array} SearchUserResult
// @Router /api/v1/search/users [get]
func SearchUsersWithQuery(c echo.Context) error {
	// 自身のユーザーIDを取得
	currentUserID, ok := c.Get("user_id").(string)
	if !ok || currentUserID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in token")
	}

	// クエリパラメータから検索条件を取得
	keyword := strings.TrimSpace(c.QueryParam("keyword"))

	// カンマ区切りの言語と国を配列に変換
	var languages []string
	languageParam := c.QueryParam("language")
	if languageParam != "" {
		for _, lang := range strings.Split(languageParam, ",") {
			if trimmed := strings.TrimSpace(lang); trimmed != "" {
				languages = append(languages, trimmed)
			}
		}
	}

	var countries []string
	countryParam := c.QueryParam("country")
	if countryParam != "" {
		for _, country := range strings.Split(countryParam, ",") {
			if trimmed := strings.TrimSpace(country); trimmed != "" {
				countries = append(countries, trimmed)
			}
		}
	}

	// DB検索実行
	results, err := db.SearchUsersAdvanced(currentUserID, keyword, languages, countries)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to search users: "+err.Error())
	}

	return c.JSON(http.StatusOK, results)
}
