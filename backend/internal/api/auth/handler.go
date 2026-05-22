package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"llm-gateway/backend/internal/service"
)

type Handler struct {
	auth *service.AuthService
}

func NewHandler(auth *service.AuthService) *Handler {
	return &Handler{auth: auth}
}

func (h *Handler) Login(c *gin.Context) {
	var input service.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		writeError(c, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	result, err := h.auth.Login(c.Request.Context(), input)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) Me(c *gin.Context) {
	token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	user, err := h.auth.CurrentUser(c.Request.Context(), token)
	if err != nil {
		handleAuthError(c, err)
		return
	}
	c.JSON(http.StatusOK, user)
}

func handleAuthError(c *gin.Context, err error) {
	var validationErr service.ValidationError
	switch {
	case errors.As(err, &validationErr):
		writeError(c, http.StatusUnauthorized, "authentication_error", validationErr.Message)
	case errors.Is(err, gorm.ErrRecordNotFound):
		writeError(c, http.StatusUnauthorized, "authentication_error", "invalid email or password")
	default:
		writeError(c, http.StatusInternalServerError, "internal_error", err.Error())
	}
}

func writeError(c *gin.Context, status int, errorType string, message string) {
	c.JSON(status, gin.H{
		"error": gin.H{
			"type":    errorType,
			"message": message,
		},
	})
}
