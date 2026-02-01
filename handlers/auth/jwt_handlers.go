package auth

import (
	"task_manager/internal/response"

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
// @Param request body response.LoginRequest true "Login request"
// @Success 200 {object} response.AuthTokenEnvelope
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Router /auth/login [post]
func Login(c *gin.Context) {
	if mw == nil {
		response.Internal(response.CodeInternalError, "auth middleware not initialized", "nil middleware", nil).Send(c)
		return
	}
	mw.LoginHandler(c)
}

// Refresh godoc
// @Summary Refresh token
// @Description Refresh JWT access token using refresh token.
// @Tags auth
// @Produce json
// @Success 200 {object} response.AuthTokenEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Router /auth/refresh [post]
func Refresh(c *gin.Context) {
	if mw == nil {
		response.Internal(response.CodeInternalError, "auth middleware not initialized", "nil middleware", nil).Send(c)
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
// @Success 200 {object} response.LogoutEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 403 {object} response.ErrorEnvelope
// @Router /auth/logout [post]
func Logout(c *gin.Context) {
	if mw == nil {
		response.Internal(response.CodeInternalError, "auth middleware not initialized", "nil middleware", nil).Send(c)
		return
	}
	mw.LogoutHandler(c)
}
