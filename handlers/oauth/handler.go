package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"task_manager/internal/config"
	"task_manager/internal/jwtauth"
	"task_manager/internal/repositories"
	"task_manager/internal/repositories/models"
	"task_manager/internal/response"
	"task_manager/internal/trace"
	"time"

	jwt "github.com/appleboy/gin-jwt/v3"
	jwtcore "github.com/appleboy/gin-jwt/v3/core"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

type Handler struct {
	users repositories.UserRepository
	mw    *jwt.GinJWTMiddleware
	state *StateStore

	googleCfg *oauth2.Config
	githubCfg *oauth2.Config

	oauthMobileDeeplinkTemplate string
	oauthWebRedirectTemplate    string
}

func New(users repositories.UserRepository, mw *jwt.GinJWTMiddleware) *Handler {
	return NewWithConfig(users, mw, config.Load())
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/google/login", h.GoogleLogin)
	rg.GET("/google/callback", h.GoogleCallback)
	rg.GET("/google/link", h.mw.MiddlewareFunc(), h.GoogleLink)
	rg.GET("/github/login", h.GitHubLogin)
	rg.GET("/github/callback", h.GitHubCallback)
	rg.GET("/github/link", h.mw.MiddlewareFunc(), h.GitHubLink)
}

func NewWithConfig(users repositories.UserRepository, mw *jwt.GinJWTMiddleware, cfg config.Config) *Handler {
	h := &Handler{
		users:                       users,
		mw:                          mw,
		state:                       NewStateStore(10 * time.Minute),
		oauthMobileDeeplinkTemplate: cfg.OAuthMobileDeeplinkTemplate,
		oauthWebRedirectTemplate:    cfg.OAuthWebRedirectTemplate,
	}

	if cfg.GoogleClientID != "" && cfg.GoogleClientSecret != "" {
		h.googleCfg = &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.PublicBaseURL + "/api/v1/auth/google/callback",
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		}
	}

	if cfg.GitHubClientID != "" && cfg.GitHubClientSecret != "" {
		h.githubCfg = &oauth2.Config{
			ClientID:     cfg.GitHubClientID,
			ClientSecret: cfg.GitHubClientSecret,
			RedirectURL:  cfg.PublicBaseURL + "/api/v1/auth/github/callback",
			Scopes:       []string{"user:email"},
			Endpoint:     github.Endpoint,
		}
	}

	return h
}

// GoogleLogin godoc
// @Summary Google OAuth login
// @Description Redirect to Google OAuth consent screen.
// @Tags auth
// @Produce json
// @Param platform query string false "Platform for OAuth flow (mobile|web)"
// @Router /auth/google/login [get]
// @Success 302 "Redirect to Google OAuth consent screen"
// @Failure 400 {object} response.EnvelopeAny{data=response.ErrorData}
func (h *Handler) GoogleLogin(c *gin.Context) {
	if h.googleCfg == nil {
		response.BadRequest(response.CodeInvalidRequest, "Google OAuth not configured", nil).Send(c)
		return
	}
	platform := normalizePlatform(c.Query("platform"))
	state := h.state.GenerateWithData(StateData{Platform: platform, Mode: "login"})
	url := h.googleCfg.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleLink godoc
// @Summary Google OAuth link
// @Description Link Google account to existing user.
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Param platform query string false "Platform for OAuth flow (mobile|web)"
// @Success 302 "Redirect to Google OAuth consent screen"
// @Failure 400 {object} response.EnvelopeAny{data=response.ErrorData}
// @Router /auth/google/link [get]
func (h *Handler) GoogleLink(c *gin.Context) {
	if h.googleCfg == nil {
		response.BadRequest(response.CodeInvalidRequest, "Google OAuth not configured", nil).Send(c)
		return
	}
	userID, err := currentUserID(c)
	if err != nil {
		response.Unauthorized(response.CodeUnauthorized, "unauthorized", nil).Send(c)
		return
	}
	platform := normalizePlatform(c.Query("platform"))
	state := h.state.GenerateWithData(StateData{Platform: platform, Mode: "link", UserID: userID.String()})
	url := h.googleCfg.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *Handler) GoogleCallback(c *gin.Context) {
	if h.googleCfg == nil {
		response.BadRequest(response.CodeInvalidRequest, "Google OAuth not configured", nil).Send(c)
		return
	}

	state := c.Query("state")
	stateData, ok := h.state.Consume(state)
	if !ok {
		response.BadRequest(response.CodeInvalidRequest, "Invalid state token", nil).Send(c)
		return
	}

	code := c.Query("code")
	tok, err := h.googleCfg.Exchange(context.Background(), code)
	if err != nil {
		response.Internal(response.CodeInternalError, "Failed to exchange token", err.Error(), nil).Send(c)
		return
	}

	client := h.googleCfg.Client(context.Background(), tok)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		response.Internal(response.CodeInternalError, "Failed to get user info", err.Error(), nil).Send(c)
		return
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	var gu struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.Unmarshal(data, &gu); err != nil {
		response.Internal(response.CodeInternalError, "Failed to parse user info", err.Error(), nil).Send(c)
		return
	}

	email := strings.ToLower(strings.TrimSpace(gu.Email))
	profile := providerProfile{
		Provider:       models.GoogleProvider,
		ProviderUserID: strings.TrimSpace(gu.ID),
		Email:          email,
		DisplayName:    strings.TrimSpace(gu.Name),
		AvatarURL:      strings.TrimSpace(gu.Picture),
	}
	h.handleProviderCallback(c, stateData, profile)
}

// GitHubLogin godoc
// @Summary GitHub OAuth login
// @Description Redirect to GitHub OAuth consent screen.
// @Tags auth
// @Produce json
// @Param platform query string false "Platform for OAuth flow (mobile|web)"
// @Success 302 "Redirect to Google OAuth consent screen"
// @Failure 400 {object} response.EnvelopeAny{data=response.ErrorData}
// @Router /auth/github/login [get]
func (h *Handler) GitHubLogin(c *gin.Context) {
	if h.githubCfg == nil {
		response.BadRequest(response.CodeInvalidRequest, "GitHub OAuth not configured", nil).Send(c)
		return
	}
	platform := normalizePlatform(c.Query("platform"))
	state := h.state.GenerateWithData(StateData{Platform: platform, Mode: "login"})
	url := h.githubCfg.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GitHubLink godoc
// @Summary GitHub OAuth link
// @Description Link GitHub account to existing user.
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Param platform query string false "Platform for OAuth flow (mobile|web)"
// @Success 302 "Redirect to Google OAuth consent screen"
// @Failure 400 {object} response.EnvelopeAny{data=response.ErrorData}
// @Router /auth/github/link [get]
func (h *Handler) GitHubLink(c *gin.Context) {
	if h.githubCfg == nil {
		response.BadRequest(response.CodeInvalidRequest, "GitHub OAuth not configured", nil).Send(c)
		return
	}
	userID, err := currentUserID(c)
	if err != nil {
		response.Unauthorized(response.CodeUnauthorized, "unauthorized", nil).Send(c)
		return
	}
	platform := normalizePlatform(c.Query("platform"))
	state := h.state.GenerateWithData(StateData{Platform: platform, Mode: "link", UserID: userID.String()})
	url := h.githubCfg.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *Handler) GitHubCallback(c *gin.Context) {
	if h.githubCfg == nil {
		response.BadRequest(response.CodeInvalidRequest, "GitHub OAuth not configured", nil).Send(c)
		return
	}

	state := c.Query("state")
	stateData, ok := h.state.Consume(state)
	if !ok {
		response.BadRequest(response.CodeInvalidRequest, "Invalid state token", nil).Send(c)
		return
	}

	code := c.Query("code")
	tok, err := h.githubCfg.Exchange(context.Background(), code)
	if err != nil {
		response.Internal(response.CodeInternalError, "Failed to exchange token", err.Error(), nil).Send(c)
		return
	}

	client := h.githubCfg.Client(context.Background(), tok)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		response.Internal(response.CodeInternalError, "Failed to get user info", err.Error(), nil).Send(c)
		return
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	var gh struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Email     string `json:"email"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.Unmarshal(data, &gh); err != nil {
		response.Internal(response.CodeInternalError, "Failed to parse user info", err.Error(), nil).Send(c)
		return
	}

	email := strings.ToLower(strings.TrimSpace(gh.Email))
	if email == "" {
		email = strings.ToLower(strings.TrimSpace(getGitHubPrimaryEmail(client)))
	}
	if email == "" {
		response.Internal(response.CodeInternalError, "Failed to get user email", "GitHub returned no email", nil).Send(c)
		return
	}

	profile := providerProfile{
		Provider:       models.GithubProvider,
		ProviderUserID: strconvItoa(gh.ID),
		Email:          email,
		Username:       strings.TrimSpace(gh.Login),
		DisplayName:    strings.TrimSpace(gh.Name),
		AvatarURL:      strings.TrimSpace(gh.AvatarURL),
	}
	h.handleProviderCallback(c, stateData, profile)
}

type providerProfile struct {
	Provider       models.Provider
	ProviderUserID string
	Email          string
	Username       string
	DisplayName    string
	AvatarURL      string
}

func (h *Handler) handleProviderCallback(c *gin.Context, stateData StateData, p providerProfile) {
	if p.Provider == "" || p.ProviderUserID == "" {
		response.BadRequest(response.CodeInvalidRequest, "invalid provider response", nil).Send(c)
		return
	}

	switch stateData.Mode {
	case "", "login":
		h.handleLoginWithProvider(c, stateData.Platform, p)
		return
	case "link":
		h.handleLinkProvider(c, stateData.Platform, stateData.UserID, p)
		return
	default:
		response.BadRequest(response.CodeInvalidRequest, "invalid state", nil).Send(c)
		return
	}
}

func (h *Handler) handleLoginWithProvider(c *gin.Context, platform string, p providerProfile) {
	// 1) If provider is already linked, login.
	existingByProvider, err := h.users.GetUserByAuthProvider(c.Request.Context(), p.Provider, p.ProviderUserID)
	if err != nil {
		response.Internal(response.CodeDatabaseError, "database error", err.Error(), nil).Send(c)
		return
	}
	if existingByProvider != nil {
		trace.Log(c, "oauth_login",
			"provider="+string(p.Provider)+
				" provider_user_id="+p.ProviderUserID+
				" user_id="+existingByProvider.ID.String()+
				" email="+existingByProvider.Email,
		)
		h.issueToken(c, &jwtauth.UserIdentity{
			ID:        existingByProvider.ID.String(),
			Email:     existingByProvider.Email,
			FirstName: existingByProvider.FirstName,
			LastName:  existingByProvider.LastName,
			UserType:  existingByProvider.UserType,
			Provider:  p.Provider,
			Avatar:    p.AvatarURL,
		}, platform)
		return
	}

	// 2) If email matches an existing user but provider isn't linked yet, forbid.
	if strings.TrimSpace(p.Email) != "" {
		existingByEmail, err := h.users.GetUserByEmail(c.Request.Context(), p.Email)
		if err != nil {
			response.Internal(response.CodeDatabaseError, "database error", err.Error(), nil).Send(c)
			return
		}
		if existingByEmail != nil {
			trace.Log(c, "oauth_login_forbidden_not_linked",
				"provider="+string(p.Provider)+" provider_user_id="+p.ProviderUserID+" email="+p.Email,
			)
			response.Fail(c, http.StatusForbidden, response.CodeForbidden,
				"provider not linked to this account; login with your existing method and link it in your profile",
				"", nil,
			)
			return
		}
	}

	// 3) Create new user + provider linkage.
	if strings.TrimSpace(p.Email) == "" {
		response.Internal(response.CodeInternalError, "failed to get user email", "provider returned no email", nil).Send(c)
		return
	}

	now := time.Now()
	u := &models.User{
		ID:        uuid.New(),
		FirstName: strings.TrimSpace(p.DisplayName),
		LastName:  "",
		Email:     strings.TrimSpace(p.Email),
		CreatedAt: now,
		UpdatedAt: now,
		UserType:  models.StandardUser,
	}
	if err := h.users.CreateUser(c.Request.Context(), u); err != nil {
		response.Internal(response.CodeDatabaseError, "could not create user", err.Error(), nil).Send(c)
		return
	}
	ap := &models.AuthProvider{
		ID:             uuid.New(),
		UserID:         u.ID,
		Provider:       p.Provider,
		ProviderUserID: p.ProviderUserID,
		Email:          p.Email,
		Username:       p.Username,
		DisplayName:    p.DisplayName,
		AvatarURL:      p.AvatarURL,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := h.users.CreateAuthProvider(c.Request.Context(), ap); err != nil {
		response.Internal(response.CodeDatabaseError, "could not link provider", err.Error(), nil).Send(c)
		return
	}
	trace.Log(c, "oauth_signup",
		"provider="+string(p.Provider)+
			" provider_user_id="+p.ProviderUserID+
			" user_id="+u.ID.String()+
			" email="+u.Email,
	)

	h.issueToken(c, &jwtauth.UserIdentity{
		ID:        u.ID.String(),
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Provider:  p.Provider,
		Avatar:    p.AvatarURL,
	}, platform)
}

func (h *Handler) handleLinkProvider(c *gin.Context, platform string, userIDRaw string, p providerProfile) {
	userID, err := uuid.Parse(strings.TrimSpace(userIDRaw))
	if err != nil {
		response.BadRequest(response.CodeInvalidRequest, "invalid user id in state", nil).Send(c)
		return
	}

	u, err := h.users.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		response.Internal(response.CodeDatabaseError, "database error", err.Error(), nil).Send(c)
		return
	}
	if u == nil {
		response.BadRequest(response.CodeInvalidRequest, "user not found", nil).Send(c)
		return
	}

	// If provider_user_id already linked to another user -> conflict.
	existingByProvider, err := h.users.GetUserByAuthProvider(c.Request.Context(), p.Provider, p.ProviderUserID)
	if err != nil {
		response.Internal(response.CodeDatabaseError, "database error", err.Error(), nil).Send(c)
		return
	}
	if existingByProvider != nil && existingByProvider.ID != userID {
		response.Conflict(response.CodeConflict, "provider account already linked to another user", nil).Send(c)
		return
	}

	// If user already has this provider linked, it must match the same provider_user_id.
	existingLink, err := h.users.GetAuthProviderByUserAndProvider(c.Request.Context(), userID, p.Provider)
	if err != nil {
		response.Internal(response.CodeDatabaseError, "database error", err.Error(), nil).Send(c)
		return
	}
	if existingLink != nil && existingLink.ProviderUserID != p.ProviderUserID {
		response.Conflict(response.CodeConflict, "this provider is already linked to a different provider account", nil).Send(c)
		return
	}
	if existingLink == nil {
		now := time.Now()
		ap := &models.AuthProvider{
			ID:             uuid.New(),
			UserID:         userID,
			Provider:       p.Provider,
			ProviderUserID: p.ProviderUserID,
			Email:          p.Email,
			Username:       p.Username,
			DisplayName:    p.DisplayName,
			AvatarURL:      p.AvatarURL,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := h.users.CreateAuthProvider(c.Request.Context(), ap); err != nil {
			response.Internal(response.CodeDatabaseError, "could not link provider", err.Error(), nil).Send(c)
			return
		}
		trace.Log(c, "oauth_link",
			"provider="+string(p.Provider)+" provider_user_id="+p.ProviderUserID+" user_id="+userID.String(),
		)
	}

	// Issue a fresh token after linking (effectively "login with this provider").
	h.issueToken(c, &jwtauth.UserIdentity{
		ID:        u.ID.String(),
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Provider:  p.Provider,
		Avatar:    p.AvatarURL,
		UserType:  u.UserType,
	}, platform)
}

func currentUserID(c *gin.Context) (uuid.UUID, error) {
	claims := jwt.ExtractClaims(c)
	raw, _ := claims[jwtauth.IdentityKey].(string)
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return uuid.Nil, errors.New("missing user id claim")
	}
	return uuid.Parse(raw)
}

func (h *Handler) issueToken(c *gin.Context, user *jwtauth.UserIdentity, platform string) {
	if h.mw == nil {
		response.Internal(response.CodeInternalError, "auth middleware not initialized", "nil middleware", nil).Send(c)
		return
	}
	c.Set(h.mw.IdentityKey, user)
	token, err := h.mw.TokenGenerator(c.Request.Context(), user)
	if err != nil {
		response.Internal(response.CodeInternalError, "Failed to generate token", err.Error(), nil).Send(c)
		return
	}
	h.mw.SetCookie(c, token.AccessToken)
	h.mw.SetRefreshTokenCookie(c, token.RefreshToken)

	// Platform-aware behavior:
	// - platform=mobile: redirect to configured deep link template
	// - platform=web: redirect to configured web redirect template
	// - otherwise: keep existing JSON response behavior
	switch platform {
	case "mobile":
		if strings.TrimSpace(h.oauthMobileDeeplinkTemplate) != "" {
			c.Redirect(http.StatusTemporaryRedirect, fillRedirectTemplate(h.oauthMobileDeeplinkTemplate, token))
			return
		}
	case "web":
		if strings.TrimSpace(h.oauthWebRedirectTemplate) != "" {
			c.Redirect(http.StatusTemporaryRedirect, fillRedirectTemplate(h.oauthWebRedirectTemplate, token))
			return
		}
	}

	if h.mw.LoginResponse != nil {
		h.mw.LoginResponse(c, token)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token":  token.AccessToken,
		"refresh_token": token.RefreshToken,
		"token_type":    token.TokenType,
		"expires_at":    token.ExpiresAt,
	})
}

func normalizePlatform(raw string) string {
	p := strings.ToLower(strings.TrimSpace(raw))
	if p == "mobile" || p == "web" {
		return p
	}
	return ""
}

func fillRedirectTemplate(tmpl string, token *jwtcore.Token) string {
	expiresAt := fmt.Sprintf("%d", token.ExpiresAt)
	repl := strings.NewReplacer(
		"{access_token}", url.QueryEscape(token.AccessToken),
		"{refresh_token}", url.QueryEscape(token.RefreshToken),
		"{token_type}", url.QueryEscape(token.TokenType),
		"{expires_at}", url.QueryEscape(expiresAt),
	)
	return repl.Replace(tmpl)
}

func getGitHubPrimaryEmail(client *http.Client) string {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	var emails []struct {
		Email   string `json:"email"`
		Primary bool   `json:"primary"`
	}
	if err := json.Unmarshal(data, &emails); err != nil {
		return ""
	}
	for _, e := range emails {
		if e.Primary {
			return e.Email
		}
	}
	if len(emails) > 0 {
		return emails[0].Email
	}
	return ""
}

func strconvItoa(n int) string {
	// local helper to avoid pulling strconv for a single use in this file
	return fmt.Sprintf("%d", n)
}
