package domain

import "appa_subscriptions/internal/models"

type WebhookService interface {
	OrderCreated(webhook models.Webhook)
}
