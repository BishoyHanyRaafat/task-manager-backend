package repositories

import (
	"context"
	"task_manager/public/repositories/models"

	"github.com/google/uuid"
)

type UserRepository interface {
	CreateUser(ctx context.Context, u *models.User) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)

	// Passwords (local auth)
	UpsertPassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
	GetPasswordHashByUserID(ctx context.Context, userID uuid.UUID) (string, error)

	// OAuth providers (google/github)
	GetUserByAuthProvider(ctx context.Context, provider models.Provider, providerUserID string) (*models.User, error)
	GetAuthProviderByUserAndProvider(ctx context.Context, userID uuid.UUID, provider models.Provider) (*models.AuthProvider, error)
	ListAuthProvidersByUserID(ctx context.Context, userID uuid.UUID) ([]models.AuthProvider, error)
	CreateAuthProvider(ctx context.Context, ap *models.AuthProvider) error
}

type TeamRepository interface {
	CreateTeam(ctx context.Context, team *models.Team) error
	GetTeamByID(ctx context.Context, teamID uuid.UUID) (*models.Team, error)
	GetTeamsMembers(ctx context.Context, teamID uuid.UUID) ([]*models.UserTeam, error)
	GetTeamFounderByTeamID(ctx context.Context, teamID uuid.UUID) (*models.UserTeam, error)
	GetMemberRole(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) (*models.TeamUserRole, error)
	CreateTeamUser(ctx context.Context, userTeam *models.UserTeam) error
	DeleteTeamUser(ctx context.Context, userTeamID *uuid.UUID) error
	EditTeamName(ctx context.Context, team *models.Team) error
	DeleteTeam(ctx context.Context, teamID uuid.UUID) error
	RemoveTeamUser(ctx context.Context, userID uuid.UUID) error
	GetTeamsByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Team, error)
	CreateTeamInvitation(ctx context.Context, invitation *models.Invitation) error
	UpdateTeamInvitationStatus(ctx context.Context, invitationID uuid.UUID, accept bool) error
	GetUserInvitations(ctx context.Context, userID uuid.UUID) ([]*models.Invitation, error)
	DeleteUserInvitation(ctx context.Context, invitationID uuid.UUID) error
}
