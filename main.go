package main

import (
	"fmt"
	"task_manager/handlers/self"
	"task_manager/public/config"
	"task_manager/public/db"
	"task_manager/public/dto"
	"task_manager/public/jwtauth"
	"task_manager/public/logging"
	"task_manager/public/repositories"
	"task_manager/public/trace"
	"task_manager/public/validation"

	authhandler "task_manager/handlers/auth"
	oauthhandler "task_manager/handlers/oauth"
	userhandler "task_manager/handlers/user"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "modernc.org/sqlite"
)

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

	fmt.Println("VERSION:", config.AppVersion)
	fmt.Println("COMMIT: ", config.AppCommit)
	fmt.Println("BUILT:  ", config.BuildTime)
	_ = godotenv.Load(".env", "env.example")

	cfg := config.Load()

	loggingRes, err := logging.Init(cfg.LogFile)
	if err != nil {
		panic(err)
	}
	defer func() { _ = loggingRes.Close() }()

	r := gin.Default()

	ginconfig := cors.DefaultConfig()
	ginconfig.AllowOrigins = []string{cfg.FrontendURL}
	ginconfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE"}
	ginconfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", trace.HeaderKey}

	r.Use(cors.New(ginconfig))
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
	uow, err := repositories.NewUnitOfWork(cfg.DBDriver, sqlDB)
	if err != nil {
		panic(err)
	}
	usersRepo := uow.Users()

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
	authH := authhandler.NewHandler(uow, authMiddleware)
	oauthH := oauthhandler.NewWithConfig(uow, authMiddleware, cfg)

	// Auth routes
	authGroup := v1.Group("/auth")
	authH.RegisterRoutes(authGroup)
	authhandler.RegisterRoutes(authGroup)
	oauthH.RegisterRoutes(authGroup)

	// Self routes
	selfGroup := v1.Group("/self")
	self.RegisterRoutes(selfGroup)

	// User routes
	userGroup := v1.Group("/user")
	userhandler.RegisterRoutes(userGroup, authMiddleware.MiddlewareFunc())

	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, dto.NotFound(dto.CodeNotFound, "page not found", "DEFAULT_PAGE_HANDLER", nil))
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
