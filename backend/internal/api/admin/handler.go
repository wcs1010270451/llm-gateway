package admin

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"llm-gateway/backend/internal/api/apierror"
	"llm-gateway/backend/internal/service"
)

type Handler struct {
	providers *service.ProviderService
	models    *service.ModelService
	families  *service.ModelFamilyService
	users     *service.UserService
	logs      *service.RequestLogService
	stats     *service.AdminStatsService
}

func NewHandler(providers *service.ProviderService, models *service.ModelService, families *service.ModelFamilyService, users *service.UserService, logs *service.RequestLogService, stats *service.AdminStatsService) *Handler {
	return &Handler{
		providers: providers,
		models:    models,
		families:  families,
		users:     users,
		logs:      logs,
		stats:     stats,
	}
}

func parseID(c *gin.Context, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		apierror.BadRequest(c, "invalid id")
		return 0, false
	}
	return id, true
}

func handleServiceError(c *gin.Context, err error) {
	var validationErr service.ValidationError
	if errors.As(err, &validationErr) {
		apierror.BadRequest(c, validationErr.Message)
		return
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		apierror.NotFound(c, "record not found")
		return
	}
	apierror.Internal(c, err)
}
