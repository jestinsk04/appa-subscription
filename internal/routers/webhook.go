package routers

import (
	"appa_subscriptions/internal/handlers"
	"appa_subscriptions/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// WebhookRoutes defines the routes for the webhook service
type WebhookRoutes struct {
	handler *handlers.WebhookHandler
}

// NewWebhookRoutes creates a new instance of WebhookRoutes
func NewWebhookRoutes(
	handler *handlers.WebhookHandler,
) *WebhookRoutes {
	return &WebhookRoutes{
		handler: handler,
	}
}

// SetRouter sets up the routes for the webhook service
func (r *WebhookRoutes) SetRouter(router *gin.Engine, secretKey string) {
	router.POST("/webhook/order-created", middleware.ValidateHMAC(secretKey, "X-Shopify-Hmac-Sha256"), r.handler.HandleWebhookOrderCreated)
}
