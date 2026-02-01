package config

import (
	"os"
	"strings"
)

func getEnv(key string, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

var (
	AppVersion = "dev"
	AppCommit  = "none"
	BuildTime  = "unknown"
)

type Config struct {
	Port          string
	PublicBaseURL string

	DBDriver string // sqlite | postgres
	DBDSN    string

	JWTSecret string

	LogFile string // empty disables file logging

	GoogleClientID     string
	GoogleClientSecret string
	GitHubClientID     string
	GitHubClientSecret string

	OAuthMobileDeeplinkTemplate string
	OAuthWebRedirectTemplate    string

	FrontendURL string
}

func Load() Config {
	port := getEnv("PORT", "8080")
	publicBaseURL := getEnv("PUBLIC_BASE_URL", "http://localhost:"+port)
	publicBaseURL = strings.TrimRight(publicBaseURL, "/")

	return Config{
		Port:                        port,
		PublicBaseURL:               publicBaseURL,
		DBDriver:                    getEnv("DB_DRIVER", "sqlite"),
		DBDSN:                       getEnv("DB_DSN", "file:task_manager.db?_pragma=foreign_keys(1)"),
		JWTSecret:                   getEnv("JWT_SECRET", "dev-secret-change-me"),
		LogFile:                     getEnv("LOG_FILE", "logs/app.log"),
		GoogleClientID:              getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:          getEnv("GOOGLE_CLIENT_SECRET", ""),
		GitHubClientID:              getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret:          getEnv("GITHUB_CLIENT_SECRET", ""),
		OAuthMobileDeeplinkTemplate: getEnv("OAUTH_MOBILE_DEEPLINK_TEMPLATE", ""),
		OAuthWebRedirectTemplate:    getEnv("OAUTH_WEB_REDIRECT_TEMPLATE", ""),
		FrontendURL:                 getEnv("FRONTEND_URL", "http://localhost:3000"),
	}
}
