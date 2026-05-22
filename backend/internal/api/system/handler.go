package system

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"llm-gateway/backend/internal/api/apierror"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *Handler) Ready(c *gin.Context) {
	sqlDB, err := h.db.DB()
	if err != nil {
		apierror.ServiceUnavailable(c, err)
		return
	}
	if err := sqlDB.PingContext(c.Request.Context()); err != nil {
		apierror.ServiceUnavailable(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
