package user

import (
	"net/http"
	"task_manager/internal/dto"
	"task_manager/internal/jwtauth"
	"task_manager/internal/repositories/models"

	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, AuthMiddleware gin.HandlerFunc) {
	rg.GET("/me", AuthMiddleware, Me)
}

// Me godoc
// @Summary Get current user
// @Description Returns the user_id and email from the JWT claims.
// @Tags user
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.MeEnvelope
// @Failure 401 {object} dto.ErrorEnvelope
// @Failure 403 {object} dto.ErrorEnvelope
// @Router /user/me [get]
func Me(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	id, _ := claims[jwtauth.IdentityKey].(string)
	email, _ := claims["email"].(string)
	first, _ := claims["firstname"].(string)
	last, _ := claims["lastname"].(string)
	userTypeStr, _ := claims["user_type"].(string)
	user_type := models.UserType(userTypeStr)
	dto.OK(c, http.StatusOK, dto.MeResponse{
		UserID:    id,
		FirstName: first,
		LastName:  last,
		Email:     email,
		UserType:  user_type,
	})
}
