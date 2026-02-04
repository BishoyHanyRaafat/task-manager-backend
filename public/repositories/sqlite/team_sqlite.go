package sqlite

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
		`INSERT INTO teams (id, name, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		team.ID.String(),
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
	var id string
	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, name, created_at, updated_at FROM teams WHERE id = ?`,
		teamID.String(),
	).Scan(&id, &u.Name, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	u.ID = parsed
	return &u, nil
}

func (r *TeamRepository) CreateTeamUser(ctx context.Context, userTeam *models.UserTeam) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO teams_users (id, team_id, user_id, role) VALUES (?, ?, ?, ?)`,
		userTeam.ID.String(),
		userTeam.TeamID.String(),
		userTeam.UserID.String(),
		string(userTeam.Role),
	)

	return err
}

func (r *TeamRepository) DeleteTeamUser(ctx context.Context, userTeamID *uuid.UUID) error {
	_, err := r.db.ExecContext(
		ctx,
		`DELETE FROM teams_users WHERE id = ?`,
		userTeamID.String(),
	)
	return err
}

func (r *TeamRepository) EditTeamName(ctx context.Context, team *models.Team) error {
	now := time.Now()
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE teams SET name = ?, updated_at = ? WHERE id = ?`,
		team.Name,
		now,
		team.ID.String(),
	)
	team.UpdatedAt = now
	return err
}

func (r *TeamRepository) DeleteTeam(ctx context.Context, teamID uuid.UUID) error {
	_, err := r.db.ExecContext(
		ctx,
		`DELETE FROM teams WHERE id = ?`,
		teamID.String(),
	)
	return err
}

func (r *TeamRepository) RemoveTeamUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(
		ctx,
		`DELETE FROM teams_users WHERE user_id = ?`,
		userID.String(),
	)
	return err
}

func (r *TeamRepository) GetTeamsByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Team, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT t.id, t.name, t.created_at, t.updated_at 
		 FROM teams t
		 JOIN teams_users tu ON t.id = tu.team_id
		 WHERE tu.user_id = ?
		 ORDER BY t.created_at ASC`,
		userID.String(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []*models.Team
	for rows.Next() {
		var t models.Team
		var id string
		if err := rows.Scan(&id, &t.Name, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		parsed, err := uuid.Parse(id)
		if err != nil {
			return nil, err
		}
		t.ID = parsed
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
		`INSERT INTO team_invitations (id, team_id, user_id, accepted) VALUES (?, ?, ?, ?)`,
		invitation.ID.String(),
		invitation.TeamID.String(),
		invitation.UserID.String(),
		invitation.Accepted,
	)
	return err
}

func (r *TeamRepository) UpdateTeamInvitationStatus(ctx context.Context, invitationID uuid.UUID, accept bool) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE team_invitations SET accepted = ? WHERE id = ?`,
		accept,
		invitationID.String(),
	)
	return err
}

func (r *TeamRepository) GetUserInvitations(ctx context.Context, userID uuid.UUID) ([]*models.Invitation, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, team_id, user_id, accepted FROM team_invitations WHERE user_id = ? ORDER BY id ASC`,
		userID.String(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invitations []*models.Invitation
	for rows.Next() {
		var inv models.Invitation
		var id, teamID, uID string
		if err := rows.Scan(&id, &teamID, &uID, &inv.Accepted); err != nil {
			return nil, err
		}
		parsedID, err := uuid.Parse(id)
		if err != nil {
			return nil, err
		}
		parsedTeamID, err := uuid.Parse(teamID)
		if err != nil {
			return nil, err
		}
		parsedUserID, err := uuid.Parse(uID)
		if err != nil {
			return nil, err
		}
		inv.ID = parsedID
		inv.TeamID = parsedTeamID
		inv.UserID = parsedUserID
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
		`SELECT role FROM teams_users WHERE team_id = ? AND user_id = ?`,
		teamID.String(),
		userID.String(),
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
	var id, tID, uID, role string
	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, team_id, user_id, role FROM teams_users WHERE team_id = ? AND role = ?`,
		teamID.String(),
		string(models.FounderUserRole),
	).Scan(&id, &tID, &uID, &role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	parsedTeamID, err := uuid.Parse(tID)
	if err != nil {
		return nil, err
	}
	parsedUserID, err := uuid.Parse(uID)
	if err != nil {
		return nil, err
	}
	ut.ID = parsedID
	ut.TeamID = parsedTeamID
	ut.UserID = parsedUserID
	ut.Role = models.TeamUserRole(role)
	return &ut, nil
}

func (r *TeamRepository) GetTeamsMembers(ctx context.Context, teamID uuid.UUID) ([]*models.UserTeam, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, team_id, user_id, role FROM teams_users WHERE team_id = ? ORDER BY id ASC`,
		teamID.String(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*models.UserTeam
	for rows.Next() {
		var ut models.UserTeam
		var id, tID, uID, role string
		if err := rows.Scan(&id, &tID, &uID, &role); err != nil {
			return nil, err
		}
		parsedID, err := uuid.Parse(id)
		if err != nil {
			return nil, err
		}
		parsedTeamID, err := uuid.Parse(tID)
		if err != nil {
			return nil, err
		}
		parsedUserID, err := uuid.Parse(uID)
		if err != nil {
			return nil, err
		}
		ut.ID = parsedID
		ut.TeamID = parsedTeamID
		ut.UserID = parsedUserID
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
		`DELETE FROM team_invitations WHERE id = ?`,
		invitationID.String(),
	)
	return err
}
