package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"llm-gateway/backend/internal/entity"
	"llm-gateway/backend/internal/service"
)

func requireLogin(auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := authenticateRequest(c, auth)
		if !ok {
			return
		}
		c.Set("current_user", user)
		c.Next()
	}
}

func requireRole(auth *service.AuthService, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := authenticateRequest(c, auth)
		if !ok {
			return
		}
		if user.Role != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{"type": "forbidden", "message": "permission denied"},
			})
			return
		}
		c.Set("current_user", user)
		c.Next()
	}
}

func authenticateRequest(c *gin.Context, auth *service.AuthService) (entity.User, bool) {
	token := bearerToken(c.GetHeader("Authorization"))
	if token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"type": "authentication_error", "message": "login required"},
		})
		return entity.User{}, false
	}

	user, err := auth.CurrentUser(c.Request.Context(), token)
	if err != nil {
		var validationErr service.ValidationError
		message := "invalid token"
		if errors.As(err, &validationErr) {
			message = validationErr.Message
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			message = "invalid token"
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"type": "authentication_error", "message": message},
		})
		return entity.User{}, false
	}
	return user, true
}

func bearerToken(value string) string {
	if strings.HasPrefix(strings.ToLower(value), "bearer ") {
		return strings.TrimSpace(value[7:])
	}
	return ""
}
