package apierror

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": gin.H{
			"type":    "bad_request",
			"message": message,
		},
	})
}

func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, gin.H{
		"error": gin.H{
			"type":    "not_found",
			"message": message,
		},
	})
}

func Conflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, gin.H{
		"error": gin.H{
			"type":    "conflict",
			"message": message,
		},
	})
}

func Internal(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": gin.H{
			"type":    "internal_error",
			"message": err.Error(),
		},
	})
}

func Database(c *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		NotFound(c, "record not found")
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{
		"error": gin.H{
			"type":    "database_error",
			"message": err.Error(),
		},
	})
}

func ServiceUnavailable(c *gin.Context, err error) {
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"error": gin.H{
			"type":    "service_unavailable",
			"message": err.Error(),
		},
	})
}
