package postgress

import (
	"context"
	"database/sql"
	"errors"
	dbx "task_manager/public/db"
	"task_manager/public/repositories/models"
	"time"

	"github.com/google/uuid"
)

type TeamRepository struct {
	db dbx.DBTX
}

func NewTeamRepository(db dbx.DBTX) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) CreateTeam(ctx context.Context, team *models.Team) error {
	now := time.Now()
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO teams (id, name, created_at, updated_at) VALUES ($1, $2, $3, $4)`,
		team.ID,
		team.Name,
		now,
		now,
	)
	team.CreatedAt = now
	team.UpdatedAt = now

	return err
}

func (r *TeamRepository) GetTeamByID(ctx context.Context, teamID uuid.UUID) (*models.Team, error) {
	var u models.Team
	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, name, created_at, updated_at FROM teams WHERE id = $1`,
		teamID,
	).Scan(&u.ID, &u.Name, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *TeamRepository) CreateTeamUser(ctx context.Context, userTeam *models.UserTeam) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO teams_users (id, team_id, user_id, role) VALUES ($1, $2, $3, $4)`,
		userTeam.ID,
		userTeam.TeamID,
		userTeam.UserID,
		string(userTeam.Role),
	)

	return err
}

func (r *TeamRepository) DeleteTeamUser(ctx context.Context, userTeamID *uuid.UUID) error {
	_, err := r.db.ExecContext(
		ctx,
		`DELETE FROM teams_users WHERE id = $1`,
		userTeamID,
	)
	return err
}

func (r *TeamRepository) EditTeamName(ctx context.Context, team *models.Team) error {
	now := time.Now()
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE teams SET name = $1, updated_at = $2 WHERE id = $3`,
		team.Name,
		now,
		team.ID,
	)
	team.UpdatedAt = now
	return err
}

func (r *TeamRepository) DeleteTeam(ctx context.Context, teamID uuid.UUID) error {
	_, err := r.db.ExecContext(
		ctx,
		`DELETE FROM teams WHERE id = $1`,
		teamID,
	)
	return err
}

func (r *TeamRepository) RemoveTeamUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(
		ctx,
		`DELETE FROM teams_users WHERE user_id = $1`,
		userID,
	)
	return err
}

func (r *TeamRepository) GetTeamsByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Team, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT t.id, t.name, t.created_at, t.updated_at 
		 FROM teams t
		 JOIN teams_users tu ON t.id = tu.team_id
		 WHERE tu.user_id = $1
		 ORDER BY t.created_at ASC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []*models.Team
	for rows.Next() {
		var t models.Team
		if err := rows.Scan(&t.ID, &t.Name, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		teams = append(teams, &t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return teams, nil
}

func (r *TeamRepository) CreateTeamInvitation(ctx context.Context, invitation *models.Invitation) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO team_invitations (id, team_id, user_id, accepted) VALUES ($1, $2, $3, $4)`,
		invitation.ID,
		invitation.TeamID,
		invitation.UserID,
		invitation.Accepted,
	)
	return err
}

func (r *TeamRepository) UpdateTeamInvitationStatus(ctx context.Context, invitationID uuid.UUID, accept bool) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE team_invitations SET accepted = $1 WHERE id = $2`,
		accept,
		invitationID,
	)
	return err
}

func (r *TeamRepository) GetUserInvitations(ctx context.Context, userID uuid.UUID) ([]*models.Invitation, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, team_id, user_id, accepted FROM team_invitations WHERE user_id = $1 ORDER BY id ASC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invitations []*models.Invitation
	for rows.Next() {
		var inv models.Invitation
		if err := rows.Scan(&inv.ID, &inv.TeamID, &inv.UserID, &inv.Accepted); err != nil {
			return nil, err
		}
		invitations = append(invitations, &inv)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return invitations, nil
}

func (r *TeamRepository) GetMemberRole(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) (*models.TeamUserRole, error) {
	var role string
	err := r.db.QueryRowContext(
		ctx,
		`SELECT role FROM teams_users WHERE team_id = $1 AND user_id = $2`,
		teamID,
		userID,
	).Scan(&role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	result := models.TeamUserRole(role)
	return &result, nil
}

func (r *TeamRepository) GetTeamFounderByTeamID(ctx context.Context, teamID uuid.UUID) (*models.UserTeam, error) {
	var ut models.UserTeam
	var role string
	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, team_id, user_id, role FROM teams_users WHERE team_id = $1 AND role = $2`,
		teamID,
		string(models.FounderUserRole),
	).Scan(&ut.ID, &ut.TeamID, &ut.UserID, &role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	ut.Role = models.TeamUserRole(role)
	return &ut, nil
}

func (r *TeamRepository) GetTeamsMembers(ctx context.Context, teamID uuid.UUID) ([]*models.UserTeam, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, team_id, user_id, role FROM teams_users WHERE team_id = $1 ORDER BY id ASC`,
		teamID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*models.UserTeam
	for rows.Next() {
		var ut models.UserTeam
		var role string
		if err := rows.Scan(&ut.ID, &ut.TeamID, &ut.UserID, &role); err != nil {
			return nil, err
		}
		ut.Role = models.TeamUserRole(role)
		members = append(members, &ut)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return members, nil
}

func (r *TeamRepository) DeleteUserInvitation(ctx context.Context, invitationID uuid.UUID) error {
	_, err := r.db.ExecContext(
		ctx,
		`DELETE FROM team_invitations WHERE id = $1`,
		invitationID,
	)
	return err
}
