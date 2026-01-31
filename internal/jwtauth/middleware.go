package jwtauth

import (
	"errors"
	"net/http"
	"strings"
	"task_manager/internal/repositories"
	"task_manager/internal/repositories/models"
	"task_manager/internal/response"
	"task_manager/internal/trace"
	"task_manager/internal/validation"
	"time"

	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/appleboy/gin-jwt/v3/core"
	"github.com/gin-gonic/gin"
	gojwt "github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const IdentityKey = "id"

type UserIdentity struct {
	ID        string          `json:"id"`
	Email     string          `json:"email"`
	Provider  models.Provider `json:"provider"`
	FirstName string          `json:"firstname,omitempty"`
	LastName  string          `json:"lastname,omitempty"`
	Avatar    string          `json:"avatar_url,omitempty"`
}

func New(users repositories.UserRepository, secret string) (*jwt.GinJWTMiddleware, error) {
	if secret == "" {
		return nil, errors.New("JWT secret is required")
	}
	if gin.Mode() == gin.ReleaseMode && secret == "dev-secret-change-me" {
		return nil, errors.New("refusing to start in release mode with default JWT secret; set JWT_SECRET")
	}

	return jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "task-manager",
		Key:         []byte(secret),
		Timeout:     time.Hour,
		MaxRefresh:  24 * time.Hour,
		IdentityKey: IdentityKey,

		PayloadFunc: func(data any) gojwt.MapClaims {
			if v, ok := data.(*UserIdentity); ok {
				return gojwt.MapClaims{
					IdentityKey: v.ID,
					"email":     v.Email,
					"provider":  v.Provider,
					"firstname": v.FirstName,
					"lastname":  v.LastName,
					"avatar":    v.Avatar,
				}
			}
			return gojwt.MapClaims{}
		},

		IdentityHandler: func(c *gin.Context) any {
			claims := jwt.ExtractClaims(c)
			id, _ := claims[IdentityKey].(string)
			email, _ := claims["email"].(string)
			provider, _ := claims["provider"].(models.Provider)
			first, _ := claims["firstname"].(string)
			last, _ := claims["lastname"].(string)
			avatar, _ := claims["avatar"].(string)
			return &UserIdentity{
				ID:        id,
				Email:     email,
				Provider:  provider,
				FirstName: first,
				LastName:  last,
				Avatar:    avatar,
			}
		},

		Authenticator: func(c *gin.Context) (any, error) {
			var req response.LoginRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				return nil, jwt.ErrMissingLoginValues
			}
			req.Email = strings.TrimSpace(strings.ToLower(req.Email))

			if err := validation.Validate.Struct(req); err != nil {
				return nil, jwt.ErrMissingLoginValues
			}

			u, err := users.GetUserByEmail(c.Request.Context(), req.Email)
			if err != nil || u == nil {
				return nil, jwt.ErrFailedAuthentication
			}
			hash, err := users.GetPasswordHashByUserID(c.Request.Context(), u.ID)
			if err != nil || hash == "" {
				// likely an OAuth-created user without local password
				return nil, jwt.ErrFailedAuthentication
			}
			if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
				return nil, jwt.ErrFailedAuthentication
			}

			trace.Log(c, "login_local", "user_id="+u.ID.String()+" email="+u.Email)
			return &UserIdentity{
				ID:        u.ID.String(),
				Email:     u.Email,
				FirstName: u.FirstName,
				LastName:  u.LastName,
				Provider:  "local",
			}, nil
		},

		Authorizer: func(c *gin.Context, data any) bool {
			_, ok := data.(*UserIdentity)
			return ok
		},

		Unauthorized: func(c *gin.Context, code int, message string) {
			errCode := response.CodeUnauthorized
			if code == http.StatusForbidden {
				errCode = response.CodeForbidden
			}
			if strings.Contains(strings.ToLower(message), "token") {
				if code == http.StatusUnauthorized {
					errCode = response.CodeMissingToken
				} else {
					errCode = response.CodeInvalidToken
				}
			}
			response.Fail(c, code, errCode, message, "", nil)
		},

		LoginResponse: func(c *gin.Context, token *core.Token) {
			response.OK(c, http.StatusOK, gin.H{
				"access_token":  token.AccessToken,
				"token_type":    token.TokenType,
				"refresh_token": token.RefreshToken,
				"expires_at":    token.ExpiresAt,
			})
		},

		LogoutResponse: func(c *gin.Context) {
			claims := jwt.ExtractClaims(c)
			response.OK(c, http.StatusOK, gin.H{
				"message": "Successfully logged out",
				"user":    claims["email"],
			})
		},

		// Avoid query token lookup (tokens in URLs can leak via logs, referer headers, history, etc.)
		TokenLookup:   "header: Authorization, cookie: jwt",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,

		SendCookie: true,
		// In release mode, always require secure cookies.
		SecureCookie:   gin.Mode() == gin.ReleaseMode,
		CookieHTTPOnly: true,
		// Lax helps mitigate CSRF for cookie-authenticated POSTs, while still working with normal navigations.
		CookieSameSite:    http.SameSiteLaxMode,
		CookieMaxAge:      time.Hour,
		SendAuthorization: true,
	})
}
