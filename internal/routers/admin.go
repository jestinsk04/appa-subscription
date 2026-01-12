package routers

import (
	"appa_subscriptions/internal/handlers"

	"github.com/gin-gonic/gin"
)

type AdminRoutes struct {
	handler *handlers.AdminHandler
}

func NewAdminRoutes(
	handler *handlers.AdminHandler,
) *AdminRoutes {
	return &AdminRoutes{
		handler: handler,
	}
}

func (r *AdminRoutes) SetRouter(router *gin.Engine) {
	router.GET("/admin/check-email", r.handler.HandleCheckEmailExists)
}
