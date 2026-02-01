package testutil

import (
	"task_manager/public/jwtauth"
	"task_manager/public/repositories"
	"testing"

	authhandler "task_manager/handlers/auth"
	oauthhandler "task_manager/handlers/oauth"
	mehandler "task_manager/handlers/user"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func NewTestRouter(t *testing.T, users repositories.UserRepository, jwtSecret string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	authMW, err := jwtauth.New(users, jwtSecret)
	require.NoError(t, err)
	require.NoError(t, authMW.MiddlewareInit())
	authhandler.SetMiddleware(authMW)

	r := gin.New()
	r.Use(gin.Recovery())

	v1 := r.Group("/api/v1")

	authH := authhandler.NewHandler(users, authMW)
	oauthH := oauthhandler.New(users, authMW)
	authGroup := v1.Group("/auth")
	authGroup.POST("/signup", authH.Signup)
	authGroup.POST("/login", authhandler.Login)
	authGroup.POST("/refresh", authhandler.Refresh)
	authGroup.GET("/google/login", oauthH.GoogleLogin)
	authGroup.GET("/google/callback", oauthH.GoogleCallback)
	authGroup.GET("/github/login", oauthH.GitHubLogin)
	authGroup.GET("/github/callback", oauthH.GitHubCallback)

	protected := v1.Group("/")
	protected.Use(authMW.MiddlewareFunc())
	protected.GET("/me", mehandler.Me)
	protected.POST("/logout", authhandler.Logout)

	return r
}
