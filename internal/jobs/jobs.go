package jobs

import (
	"appa_subscriptions/internal/domains"
	"context"

	"go.uber.org/zap"
)

type JobHandler struct {
	service domains.OrderService
	logger  *zap.Logger
}

func NewJobHandler(service domains.OrderService, logger *zap.Logger) *JobHandler {
	return &JobHandler{service: service, logger: logger}
}

// HandleScheduledOrders handles the scheduling of orders
func (h *JobHandler) HandleScheduledOrders() {
	if err := h.service.NextPaymentInstallmentCreate(context.Background()); err != nil {
		h.logger.Error("failed to schedule orders", zap.Error(err))
		return
	}
}
