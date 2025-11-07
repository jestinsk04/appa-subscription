package handlers

import (
	"appa_subscriptions/internal/domains"
	"appa_subscriptions/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type WebhookHandler struct {
	webhookService domains.WebhookService
}

func NewWebhookHandler(webhookService domains.WebhookService) *WebhookHandler {
	return &WebhookHandler{
		webhookService: webhookService,
	}
}

// HandleWebhookOrderCreated handles the order created webhook
func (h *WebhookHandler) HandleWebhookOrderCreated(c *gin.Context) {
	var webhook models.Webhook
	if err := c.ShouldBindJSON(&webhook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	go h.webhookService.OrderCreated(webhook)
	c.Status(http.StatusOK)
}

// HandleWebhookOrderPaid handles the order paid webhook
func (h *WebhookHandler) HandleWebhookOrderPaid(c *gin.Context) {
	var webhook models.Webhook
	if err := c.ShouldBindJSON(&webhook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	go h.webhookService.OrderPaid(webhook)
	c.Status(http.StatusOK)
}
