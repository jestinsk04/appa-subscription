package domains

import (
	"appa_subscriptions/internal/models"
	"context"
)

type WebhookService interface {
	OrderCreated(webhook models.Webhook)
	OrderPaid(webhook models.Webhook)
}

type OrderService interface {
	NextPaymentInstallmentCreate(ctx context.Context) error
}
