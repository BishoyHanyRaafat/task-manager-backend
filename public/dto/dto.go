package dto

import "task_manager/public/repositories/models"

// Request DTOs
type SignupRequest struct {
	FirstName string `json:"firstname" validate:"required,min=2,max=100"`
	LastName  string `json:"lastname" validate:"required,min=2,max=100"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type TeamCreationRequest struct {
	TeamName string `json:"team_name" validate:"required,min=2,max=100"`
}

// Response DTOs
type AuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    string `json:"expires_at"`
}

type MeResponse struct {
	UserID    string          `json:"user_id"`
	FirstName string          `json:"firstname"`
	LastName  string          `json:"lastname"`
	Email     string          `json:"email"`
	UserType  models.UserType `json:"user_type"`
}

type LogoutResponse struct {
	Message string `json:"message"`
	User    string `json:"user"`
}
type AppInfoResponse struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`
}
