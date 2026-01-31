package models

import (
	"time"

	"github.com/google/uuid"
)

//swagger:enum userType
type UserType string

const (
	StandardUser UserType = "standard"
	AdminUser    UserType = "admin"
)

func (ut UserType) IsValid() bool {
	switch ut {
	case StandardUser, AdminUser:
		return true
	default:
		return false
	}
}

type User struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	UserType  UserType  `json:"user_type"` // "admin" | "standard"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AuthProvider struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	Provider       Provider  `json:"provider"` // "google" | "github"
	ProviderUserID string    `json:"provider_user_id"`
	Email          string    `json:"email,omitempty"`
	Username       string    `json:"username,omitempty"`
	DisplayName    string    `json:"display_name,omitempty"`
	AvatarURL      string    `json:"avatar_url,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

//swagger:enum provider
type Provider string

func (p Provider) IsValid() bool {
	switch p {
	case GoogleProvider, GithubProvider, LocalProvider:
		return true
	default:
		return false
	}
}

const (
	GoogleProvider Provider = "google"
	GithubProvider Provider = "github"
	LocalProvider  Provider = "local"
)
