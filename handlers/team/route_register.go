package team

import (
	"task_manager/public/repositories"

	"github.com/gin-gonic/gin"
)

type TeamsHandler struct {
	uow repositories.UnitOfWork
}

func (r *TeamsHandler) RegisterRoutes(rg *gin.RouterGroup, AuthMiddleware gin.HandlerFunc) {
	rg.Use(AuthMiddleware)
	rg.GET("/", r.TeamGetByUserID)
	rg.POST("/", r.TeamPost)
	rg.DELETE("/:id", r.TeamDelete)
	rg.GET("/:id", r.TeamGetByID)
	rg.PUT("/:id", r.TeamEdit)

	// Teams members routes
	// rg.GET("/:id/members", r.TeamGetMembers)
	// rg.POST("/:id/members", r.TeamAddMember)
	// rg.DELETE("/:id/members/:user_id", r.TeamRemoveMember)
	// rg.PUT("/:id/members/:user_id", r.TeamEditMemberRole)

	// Invitations routes
	// rg.GET("/invitations", r.TeamGetInvitations)
	// rg.POST("/invitations", r.TeamInviteMember)
	// rg.DELETE("/invitations", r.TeamInviteDelete)
	// rg.POST("/invitations/accept", r.TeamInviteAccept)
}
