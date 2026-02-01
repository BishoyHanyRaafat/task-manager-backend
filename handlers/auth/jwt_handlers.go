package auth

import (
	"task_manager/internal/dto"

	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
)

var mw *jwt.GinJWTMiddleware

// SetMiddleware wires the gin-jwt middleware into the auth handlers.
func SetMiddleware(m *jwt.GinJWTMiddleware) {
	mw = m
}

func RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/login", Login)
	rg.POST("/refresh", Refresh)
	rg.POST("/logout", mw.MiddlewareFunc(), Logout)
}

// Login godoc
// @Summary Login
// @Description Login using email/password and return JWT access/refresh tokens.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login request"
// @Success 200 {object} dto.AuthTokenEnvelope
// @Failure 400 {object} dto.ErrorEnvelope
// @Failure 401 {object} dto.ErrorEnvelope
// @Router /auth/login [post]
func Login(c *gin.Context) {
	if mw == nil {
		dto.Internal(dto.CodeInternalError, "auth middleware not initialized", "nil middleware", nil).Send(c)
		return
	}
	mw.LoginHandler(c)
}

// Refresh godoc
// @Summary Refresh token
// @Description Refresh JWT access token using refresh token.
// @Tags auth
// @Produce json
// @Success 200 {object} dto.AuthTokenEnvelope
// @Failure 401 {object} dto.ErrorEnvelope
// @Router /auth/refresh [post]
func Refresh(c *gin.Context) {
	if mw == nil {
		dto.Internal(dto.CodeInternalError, "auth middleware not initialized", "nil middleware", nil).Send(c)
		return
	}
	mw.RefreshHandler(c)
}

// Logout godoc
// @Summary Logout
// @Description Logout (clears JWT cookies if enabled).
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.LogoutEnvelope
// @Failure 401 {object} dto.ErrorEnvelope
// @Failure 403 {object} dto.ErrorEnvelope
// @Router /auth/logout [post]
func Logout(c *gin.Context) {
	if mw == nil {
		dto.Internal(dto.CodeInternalError, "auth middleware not initialized", "nil middleware", nil).Send(c)
		return
	}
	mw.LogoutHandler(c)
}
