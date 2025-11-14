package models

import "time"

type User struct {
	ID                string     `json:"id"`
	Email             string     `json:"email"`
	Username          string     `json:"username"`
	IsOICVerified     bool       `json:"is_oic_verified"`
	CreatedAt         time.Time  `json:"created_at"`
	Major             string     `json:"major,omitempty"`
	Gender            string     `json:"gender,omitempty"`
	NativeLanguage    string     `json:"native_language,omitempty"`
	SpokenLanguages   []string   `json:"spoken_languages,omitempty"`
	LearningLanguages []string   `json:"learning_languages,omitempty"`
	Residence         string     `json:"residence,omitempty"`
	Comment           string     `json:"comment,omitempty"`
	Interests         []Interest `json:"interests,omitempty"`
	LastUpdatedAt     time.Time  `json:"last_updated,omitempty"`
}

type Interest struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	PreferenceLevel int    `json:"preference_level,omitempty"`
	Category        string `json:"category,omitempty"`
}

type UserProfileResponse struct {
	UserID            string     `json:"user_id"`
	Email             string     `json:"email"`
	Username          string     `json:"username"`
	Major             string     `json:"major"`
	Gender            string     `json:"gender"`
	NativeLanguage    string     `json:"native_language"`
	SpokenLanguages   []string   `json:"spoken_languages"`
	LearningLanguages []string   `json:"learning_languages"`
	Residence         string     `json:"residence"`
	Comment           string     `json:"comment"`
	Interests         []Interest `json:"interests"`
	LastUpdated       time.Time  `json:"last_updated"`
}

type RegisterUserRequest struct {
	Username string `json:"username"`
}

type UpdateUserProfileRequest struct {
	Username          string   `json:"username"`
	Major             string   `json:"major"`
	Gender            string   `json:"gender"`
	NativeLanguage    string   `json:"native_language"`
	SpokenLanguages   []string `json:"spoken_languages"`
	LearningLanguages []string `json:"learning_languages"`
	Residence         string   `json:"residence"`
	Comment           string   `json:"comment"`
	InterestIDs       []int    `json:"interest_ids"`
}
