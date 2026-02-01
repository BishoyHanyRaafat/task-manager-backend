package user

import (
	"net/http"
	"task_manager/internal/jwtauth"
	"task_manager/internal/repositories/models"
	"task_manager/internal/response"

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
// @Success 200 {object} response.EnvelopeAny{data=response.MeData}
// @Failure 401 {object} response.EnvelopeAny{data=response.ErrorData}
// @Failure 403 {object} response.EnvelopeAny{data=response.ErrorData}
// @Router /me [get]
func Me(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	id, _ := claims[jwtauth.IdentityKey].(string)
	email, _ := claims["email"].(string)
	first, _ := claims["firstname"].(string)
	last, _ := claims["lastname"].(string)
	userTypeStr, _ := claims["user_type"].(string)
	user_type := models.UserType(userTypeStr)
	response.OK(c, http.StatusOK, response.MeData{
		UserID:    id,
		FirstName: first,
		LastName:  last,
		Email:     email,
		UserType:  user_type,
	})
}
