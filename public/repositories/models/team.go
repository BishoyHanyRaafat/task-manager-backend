package models

import (
	"time"

	"github.com/google/uuid"
)

type Task struct {
	ID         uuid.UUID  `json:"id"`
	TeamID     uuid.UUID  `json:"team_id"`
	Grouped    bool       `json:"group_task"`
	UserTaskID *uuid.UUID `json:"user_task_id"`
}

// TODO: Split into two models
type Team struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserTeam struct {
	ID     uuid.UUID    `json:"id"`
	TeamID uuid.UUID    `json:"team_id"`
	UserID uuid.UUID    `json:"user_id"`
	Role   TeamUserRole `json:"role"`
}

type Invitation struct {
	ID     uuid.UUID `json:"id"`
	TeamID uuid.UUID `json:"team_id"`
	UserID uuid.UUID `json:"user_id"`
	// Can be None means it is not accepted yet
	Accepted *bool `json:"accepted"`
}

//swagger:enum TeamUserRole
type TeamUserRole string

func (p TeamUserRole) IsValid() bool {
	switch p {
	case StandardUserRole, AdminUserRole, FounderUserRole:
		return true
	default:
		return false
	}
}

const (
	StandardUserRole TeamUserRole = "standard"
	AdminUserRole    TeamUserRole = "admin"
	FounderUserRole  TeamUserRole = "founder"
)
