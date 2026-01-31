package main

import (
	"fmt"
	"task_manager/internal/config"
	"task_manager/internal/db"
	"task_manager/internal/jwtauth"
	"task_manager/internal/logging"
	"task_manager/internal/repositories"
	"task_manager/internal/response"
	"task_manager/internal/trace"
	"task_manager/internal/validation"

	authhandler "task_manager/handlers/auth"
	mehandler "task_manager/handlers/me"
	oauthhandler "task_manager/handlers/oauth"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "modernc.org/sqlite"
)

var (
	appVersion = "dev"
	appCommit  = "none"
	buildTime  = "unknown"
)

type AppInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`
}

// @title Task Manager API
// @version dev
// @host localhost:8080
// @description Task manager API
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	fmt.Println("Starting app...")
	_ = godotenv.Load(".env", "env.example")

	cfg := config.Load()

	loggingRes, err := logging.Init(cfg.LogFile)
	if err != nil {
		panic(err)
	}
	defer func() { _ = loggingRes.Close() }()

	r := gin.Default()
	r.Use(trace.Middleware())

	sqlDB, err := db.Open(db.Config{Driver: cfg.DBDriver, DSN: cfg.DBDSN})
	if err != nil {
		panic(err)
	}
	defer sqlDB.Close()
	validation.RegisterValidations()

	// Run schema migrations via migration tool (SQL files under /migrations)
	migrationsDir := "migrations/" + cfg.DBDriver
	if err := db.RunMigrations(cfg.DBDriver, sqlDB, migrationsDir); err != nil {
		panic(err)
	}

	// Repositories DB LOGIC
	usersRepo, err := repositories.NewUserRepository(cfg.DBDriver, sqlDB)
	if err != nil {
		panic(err)
	}

	// Auth Middleware config
	authMiddleware, err := jwtauth.New(usersRepo, cfg.JWTSecret)
	if err != nil {
		panic(err)
	}
	if err := authMiddleware.MiddlewareInit(); err != nil {
		panic(err)
	}
	authhandler.SetMiddleware(authMiddleware)

	v1 := r.Group("/api/v1")
	// Info route to check the health/version
	// @Summary Get application info
	// @Produce json
	// @Tag meta
	// @Success 200 {object} AppInfo
	// @Router /api/v1/info [get]
	v1.GET("/info", func(c *gin.Context) {
		c.JSON(200, AppInfo{
			Version:   appVersion,
			Commit:    appCommit,
			BuildTime: buildTime,
		})
	})
	authH := authhandler.NewHandler(usersRepo, authMiddleware)
	oauthH := oauthhandler.NewWithConfig(usersRepo, authMiddleware, cfg)

	// Auth routes
	authGroup := v1.Group("/auth")
	authH.RegisterRoutes(authGroup)
	authhandler.RegisterRoutes(authGroup)
	oauthH.RegisterRoutes(authGroup)
	mehandler.RegisterRoutes(v1, authMiddleware.MiddlewareFunc())

	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, response.NotFound(response.CodeNotFound, "page not found", "DEFAULT_PAGE_HANDLER", nil))
	})

	// Swagger/OpenAPI (debug mode only)
	if gin.Mode() != gin.ReleaseMode {
		r.GET("/openapi.json", func(c *gin.Context) {
			c.File("docs/openapi.json")
		})
		// Swagger UI:
		// - UI:  /docs/index.html
		// - Spec: /openapi.json
		r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/openapi.json")))
	}

	if err := r.Run(":" + cfg.Port); err != nil {
		fmt.Println("Failed to run server:", err)
	}
}
