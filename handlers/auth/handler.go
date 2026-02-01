package auth

import (
	"net/http"
	"strings"
	"task_manager/internal/dto"
	"task_manager/internal/jwtauth"
	"task_manager/internal/repositories"
	"task_manager/internal/repositories/models"
	"task_manager/internal/trace"
	"task_manager/internal/validation"
	"time"

	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	users repositories.UserRepository
	mw    *jwt.GinJWTMiddleware
}

func NewHandler(users repositories.UserRepository, mw *jwt.GinJWTMiddleware) *Handler {
	return &Handler{users: users, mw: mw}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/signup", h.Signup)
}

// Signup godoc
// @Summary Signup
// @Description Create a new user (email/password) and return a JWT access token.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.SignupRequest true "Signup request"
// @Success 200 {object} dto.AuthTokenEnvelope
// @Failure 400 {object} dto.ErrorEnvelope
// @Failure 409 {object} dto.ErrorEnvelope
// @Failure 500 {object} dto.ErrorEnvelope
// @Router /auth/signup [post]
func (h *Handler) Signup(c *gin.Context) {
	req := dto.SignupRequest{}

	if err := c.BindJSON(&req); err != nil {
		dto.BadRequest(dto.CodeInvalidRequest, "invalid request", nil).Send(c)
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)

	if err := validation.Validate.Struct(req); err != nil {
		msg, details, dbg := validation.Format(err)
		dto.Fail(c, http.StatusBadRequest, dto.CodeValidationError, msg, dbg, details)
		return
	}

	existing, err := h.users.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		dto.Internal(dto.CodeDatabaseError, "database error", err.Error(), nil).Send(c)
		return
	}
	if existing != nil {
		dto.Conflict(dto.CodeConflict, "email already in use", nil).Send(c)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 14)
	if err != nil {
		dto.Internal(dto.CodeInternalError, "could not hash password", err.Error(), nil).Send(c)
		return
	}

	now := time.Now()
	u := &models.User{
		ID:        uuid.New(),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		UserType:  models.StandardUser,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := h.users.CreateUser(c.Request.Context(), u); err != nil {
		// if the DB enforces uniqueness, this also covers race conditions
		if looksLikeUniqueViolation(err) {
			dto.Conflict(dto.CodeConflict, "email already in use", nil).Send(c)
			return
		}
		dto.Internal(dto.CodeDatabaseError, "could not create user", err.Error(), nil).Send(c)
		return
	}

	if err := h.users.UpsertPassword(c.Request.Context(), u.ID, string(hash)); err != nil {
		dto.Internal(dto.CodeDatabaseError, "could not set password", err.Error(), nil).Send(c)
		return
	}
	trace.Log(c, "signup", "user_id="+u.ID.String()+" email="+u.Email)

	if h.mw == nil {
		dto.Internal(dto.CodeInternalError, "auth middleware not initialized", "nil middleware", nil).Send(c)
		return
	}

	identity := &jwtauth.UserIdentity{
		ID:        u.ID.String(),
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Provider:  models.LocalProvider,
		UserType:  u.UserType,
	}
	c.Set(h.mw.IdentityKey, identity)

	token, err := h.mw.TokenGenerator(c.Request.Context(), identity)
	if err != nil {
		dto.Internal(dto.CodeInternalError, "could not generate token", err.Error(), nil).Send(c)
		return
	}
	h.mw.SetCookie(c, token.AccessToken)
	h.mw.SetRefreshTokenCookie(c, token.RefreshToken)

	if h.mw.LoginResponse != nil {
		h.mw.LoginResponse(c, token)
		return
	}
	dto.OK(c, http.StatusOK, gin.H{
		"access_token":  token.AccessToken,
		"refresh_token": token.RefreshToken,
		"token_type":    token.TokenType,
		"expires_at":    token.ExpiresAt,
	})
}

func looksLikeUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "unique") || strings.Contains(s, "duplicate")
}
