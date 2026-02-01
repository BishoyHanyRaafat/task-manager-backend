package jwtauth

import (
	"errors"
	"net/http"
	"strings"
	"task_manager/public/dto"
	"task_manager/public/repositories"
	"task_manager/public/repositories/models"
	"task_manager/public/trace"
	"task_manager/public/validation"
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
	UserType  models.UserType `json:"user_type,omitempty"`
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
					"user_type": v.UserType,
				}
			}
			return gojwt.MapClaims{}
		},

		IdentityHandler: func(c *gin.Context) any {
			claims := jwt.ExtractClaims(c)
			id, _ := claims[IdentityKey].(string)
			email, _ := claims["email"].(string)
			provider, _ := claims["provider"].(string)
			first, _ := claims["firstname"].(string)
			last, _ := claims["lastname"].(string)
			avatar, _ := claims["avatar"].(string)
			userType, _ := claims["user_type"].(models.UserType)
			return &UserIdentity{
				ID:        id,
				Email:     email,
				Provider:  models.Provider(provider),
				FirstName: first,
				LastName:  last,
				Avatar:    avatar,
				UserType:  userType,
			}
		},

		Authenticator: func(c *gin.Context) (any, error) {
			var req dto.LoginRequest
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

			trace.Log(c, "login_local", "user_id="+u.ID.String()+" email="+u.Email+" user_type="+string(u.UserType))
			return &UserIdentity{
				ID:        u.ID.String(),
				Email:     u.Email,
				FirstName: u.FirstName,
				LastName:  u.LastName,
				Provider:  models.LocalProvider,
				UserType:  u.UserType,
			}, nil
		},

		Authorizer: func(c *gin.Context, data any) bool {
			_, ok := data.(*UserIdentity)
			return ok
		},

		Unauthorized: func(c *gin.Context, code int, message string) {
			errCode := dto.CodeUnauthorized
			if code == http.StatusForbidden {
				errCode = dto.CodeForbidden
			}
			if strings.Contains(strings.ToLower(message), "token") {
				if code == http.StatusUnauthorized {
					errCode = dto.CodeMissingToken
				} else {
					errCode = dto.CodeInvalidToken
				}
			}
			dto.Fail(c, code, errCode, message, "", nil)
		},

		LoginResponse: func(c *gin.Context, token *core.Token) {
			dto.OK(c, http.StatusOK, gin.H{
				"access_token":  token.AccessToken,
				"token_type":    token.TokenType,
				"refresh_token": token.RefreshToken,
				"expires_at":    token.ExpiresAt,
			})
		},

		LogoutResponse: func(c *gin.Context) {
			claims := jwt.ExtractClaims(c)

			email, ok := claims["email"].(string)
			if !ok {
				email = ""
			}

			dto.OK(c, http.StatusOK, dto.LogoutResponse{
				Message: "Successfully logged out",
				User:    email,
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
