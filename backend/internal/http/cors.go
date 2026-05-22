package httpapi

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func corsMiddleware(frontendOrigins string) gin.HandlerFunc {
	allowedOrigins := parseAllowedOrigins(frontendOrigins)

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && allowedOrigins[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, anthropic-version, anthropic-beta")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func parseAllowedOrigins(value string) map[string]bool {
	items := strings.Split(value, ",")
	origins := make(map[string]bool, len(items))
	for _, item := range items {
		origin := strings.TrimSpace(item)
		if origin != "" {
			origins[origin] = true
		}
	}
	return origins
}
