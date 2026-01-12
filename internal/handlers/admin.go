package handlers

import (
	"appa_subscriptions/internal/domains"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	service domains.AdminService
}

// NewAdminHandler creates a new instance of AdminHandler
func NewAdminHandler(service domains.AdminService) *AdminHandler {
	return &AdminHandler{
		service: service,
	}
}

// HandleCheckEmailExists handles the email existence check
func (h *AdminHandler) HandleCheckEmailExists(c *gin.Context) {
	email := c.Query("email")
	exists, err := h.service.CheckEmailExists(email)
	if err != nil {
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(200, gin.H{"exists": exists})
}
