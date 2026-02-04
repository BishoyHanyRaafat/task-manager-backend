package team

import (
	"net/http"
	"strings"
	"task_manager/public/dto"
	"task_manager/public/jwtauth"
	"task_manager/public/repositories/models"
	"task_manager/public/validation"
	"time"

	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TeamGetByUserID godoc
// @Summary Get current teams
// @Description Get the teams that the user currently in
// @Tags teams
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.TeamsEnvelope
// @Failure 401 {object} dto.ErrorEnvelope
// @Failure 403 {object} dto.ErrorEnvelope
// @Router /team [get]
func (r *TeamsHandler) TeamGetByUserID(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	id, _ := claims[jwtauth.IdentityKey].(string)
	userID, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		dto.BadRequest(dto.CodeInvalidRequest, "invalid user id in state", nil).Send(c)
		return
	}
	teams, err := r.uow.Teams().GetTeamsByUserID(c.Request.Context(), userID)
	if err != nil {
		dto.Internal(dto.CodeInternalError, "Internal server error", err.Error(), nil).Send(c)
		return
	}
	dto.OK(c, http.StatusOK, teams)
}

// TeamPost godoc
// @Summary Create a new team
// @Tags teams
// @Produce json
// @Param request body dto.TeamCreationRequest true "Team creation request"
// @Security BearerAuth
// @Success 200 {object} dto.TeamsEnvelope
// @Failure 401 {object} dto.ErrorEnvelope
// @Failure 403 {object} dto.ErrorEnvelope
// @Router /team [post]
func (r *TeamsHandler) TeamPost(c *gin.Context) {
	claims := jwt.ExtractClaims(c)

	req := dto.TeamCreationRequest{}

	if err := c.BindJSON(&req); err != nil {
		dto.BadRequest(dto.CodeInvalidRequest, "invalid request", nil).Send(c)
		return
	}
	req.TeamName = strings.TrimSpace(req.TeamName)

	if err := validation.Validate.Struct(req); err != nil {
		msg, details, dbg := validation.Format(err)
		dto.Fail(c, http.StatusBadRequest, dto.CodeValidationError, msg, dbg, details)
		return
	}

	id, _ := claims[jwtauth.IdentityKey].(string)
	userID, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		dto.BadRequest(dto.CodeInvalidRequest, "invalid user id in state", nil).Send(c)
		return
	}

	tx, err := r.uow.Begin(c.Request.Context())
	if err != nil {
		dto.Internal(dto.CodeInternalError, "Internal server error", err.Error(), nil).Send(c)
		return
	}

	defer tx.Stop()
	now := time.Now()
	team := &models.Team{
		ID:        uuid.New(),
		Name:      req.TeamName,
		CreatedAt: now,
		UpdatedAt: now,
	}
	teamsRepo := tx.Teams()
	err = teamsRepo.CreateTeam(c.Request.Context(), team)
	if err != nil {
		dto.Internal(dto.CodeInternalError, "Internal server error", err.Error(), nil).Send(c)
		return
	}

	err = teamsRepo.CreateTeamUser(c.Request.Context(), &models.UserTeam{
		ID:     uuid.New(),
		UserID: userID,
		TeamID: team.ID,
		Role:   models.FounderUserRole,
	})
	if err != nil {
		dto.Internal(dto.CodeInternalError, "Internal server error", err.Error(), nil).Send(c)
		return
	}

	dto.OK(c, http.StatusOK, team)
}

// TeamDelete godoc
// @Summary Delete a team
// @Tags teams
// @Produce json
// @Param request body dto.TeamCreationRequest true "Team creation request"
// @Security BearerAuth
// @Param id path string true "Team ID"
// @Success 204 "No Content"
// @Failure 401 {object} dto.ErrorEnvelope
// @Failure 403 {object} dto.ErrorEnvelope
// @Router /team/{id} [delete]
func (r *TeamsHandler) TeamDelete(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	userid, _ := claims[jwtauth.IdentityKey].(string)
	id := c.Param("id")
	idUUID, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		dto.BadRequest(dto.CodeInvalidRequest, "invalid team id", nil).Send(c)
		return
	}
	founder, err := r.uow.Teams().GetTeamFounderByTeamID(c.Request.Context(), idUUID)
	if err != nil {
		dto.Internal(dto.CodeInternalError, "Internal server error", err.Error(), nil).Send(c)
		return
	}

	if founder == nil || founder.UserID.String() != strings.TrimSpace(userid) {
		dto.Forbidden(dto.CodeForbidden, "only founder can delete the team", nil).Send(c)
		return
	}

	r.uow.Teams().DeleteTeam(c.Request.Context(), idUUID)
	c.Status(http.StatusNoContent)
}

// TeamEdit godoc
// @Summary Edit team name
// @tags teams
// @Produce json
// @Param id path string true "Team ID"
// @Param request body dto.TeamCreationRequest true "Team edit request"
// @Security BearerAuth
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorEnvelope
// @Failure 401 {object} dto.ErrorEnvelope
// @Failure 403 {object} dto.ErrorEnvelope
// @Router /team/{id} [put]
func (r *TeamsHandler) TeamEdit(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	userid, _ := claims[jwtauth.IdentityKey].(string)
	teamid := c.Param("id")
	teamidUUID, err := uuid.Parse(strings.TrimSpace(teamid))
	if err != nil {
		dto.BadRequest(dto.CodeInvalidRequest, "invalid team id", nil).Send(c)
		return
	}
	getUserRole, err := r.uow.Teams().GetMemberRole(c.Request.Context(), teamidUUID, uuid.MustParse(strings.TrimSpace(userid)))
	if err != nil {
		dto.Internal(dto.CodeInternalError, "Internal server error", err.Error(), nil).Send(c)
		return
	}
	if getUserRole == nil || (*getUserRole != models.AdminUserRole && *getUserRole != models.FounderUserRole) {
		dto.Forbidden(dto.CodeForbidden, "only admin or founder can edit the team", nil).Send(c)
		return
	}

	req := dto.TeamCreationRequest{}

	if err := c.BindJSON(&req); err != nil {
		dto.BadRequest(dto.CodeInvalidRequest, "invalid request", nil).Send(c)
		return
	}
	r.uow.Teams().EditTeamName(c.Request.Context(), &models.Team{
		ID:   teamidUUID,
		Name: req.TeamName,
	})

	c.Status(http.StatusNoContent)
}

// TeamGetByID godoc
// @Summary Get team by ID
// @Param id path string true "Team ID"
// @Security BearerAuth
// @Success 200 {object} models.Team
// @Failure 400 {object} dto.ErrorEnvelope
// @Failure 401 {object} dto.ErrorEnvelope
// @Failure 403 {object} dto.ErrorEnvelope
// @Router /team/{id} [get]
func (r *TeamsHandler) TeamGetByID(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	id, _ := claims[jwtauth.IdentityKey].(string)
	TeamID := c.Param("id")
	teamIDUUID, err := uuid.Parse(strings.TrimSpace(TeamID))
	if err != nil {
		dto.BadRequest(dto.CodeInvalidRequest, "invalid team id", nil).Send(c)
		return
	}

	getUserRole, err := r.uow.Teams().GetMemberRole(c.Request.Context(), teamIDUUID, uuid.MustParse(strings.TrimSpace(id)))
	if err != nil {
		dto.Internal(dto.CodeInternalError, "Internal server error", err.Error(), nil).Send(c)
		return
	}

	if getUserRole == nil {
		dto.Forbidden(dto.CodeForbidden, "only team members can get the team info", nil).Send(c)
	}

	Team, err := r.uow.Teams().GetTeamByID(c.Request.Context(), teamIDUUID)
	if err != nil {
		dto.Internal(dto.CodeInternalError, "Internal server error", err.Error(), nil).Send(c)
	}

	dto.OK(c, http.StatusOK, Team)
}
