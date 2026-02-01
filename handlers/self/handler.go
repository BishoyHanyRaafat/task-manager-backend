package self

import (
	"task_manager/public/config"
	"task_manager/public/dto"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/info", getInfo)
	rg.GET("/health", Health)
}

// Info route to check the health/version
// @Summary Get application info
// @Description Returns the application version, commit hash, and build time.
// @Produce json
// @Tags self
// @Success 200 {object} dto.AppInfoResponse
// @Router /self/info [get]
func getInfo(c *gin.Context) {
	c.JSON(200, dto.AppInfoResponse{
		Version:   config.AppVersion,
		Commit:    config.AppCommit,
		BuildTime: config.BuildTime,
	})
}

// Health route to check the health/status
// @Summary Health check
// @Description Returns the health status of the application.
// @Produce text/plain
// @Tags self
// @Success 200 {string} string "OK"
// @Router /self/health [get]
func Health(c *gin.Context) {
	dto.OK(c, 200, "OK")
}
