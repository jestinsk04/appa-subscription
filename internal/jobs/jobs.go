package jobs

import (
	"appa_subscriptions/internal/domains"
	"context"

	"go.uber.org/zap"
)

type JobHandler struct {
	ordersService domains.OrderService
	logger        *zap.Logger
}

func NewJobHandler(ordersService domains.OrderService, logger *zap.Logger) *JobHandler {
	return &JobHandler{ordersService: ordersService, logger: logger}
}

// HandleScheduledOrders handles the scheduling of orders
func (h *JobHandler) HandleScheduledOrders() {
	if err := h.ordersService.NextPaymentInstallmentCreate(context.Background()); err != nil {
		h.logger.Error("failed to schedule orders", zap.Error(err))
		return
	}
}

// HandleReminderPendingPolicies handles sending reminders for pending policies
func (h *JobHandler) HandleReminderPendingPolicies() {
	if err := h.ordersService.ReminderPendingPolicies(context.Background()); err != nil {
		h.logger.Error("failed to send reminders for pending policies", zap.Error(err))
		return
	}
}
