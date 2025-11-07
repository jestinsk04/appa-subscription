package PaymentInstallment

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	dbModels "appa_subscriptions/pkg/db/models"
)

type Repository interface {
	Create(
		tx *gorm.DB,
		ctx context.Context,
		orderID string,
		status string,
		orderName, amountStr string,
	) (*dbModels.PaymentInstallment, error)
}

type repository struct {
	loc    *time.Location
	logger *zap.Logger
}

func NewPaymentInstallmentRepository(
	loc *time.Location,
	logger *zap.Logger,
) Repository {
	return &repository{
		loc:    loc,
		logger: logger,
	}
}

// Create a new payment installment
func (r *repository) Create(
	tx *gorm.DB,
	ctx context.Context,
	orderID string,
	status string,
	orderName, amountStr string,
) (*dbModels.PaymentInstallment, error) {
	if tx == nil {
		r.logger.Error("transaction is nil")
		return nil, fmt.Errorf("transaction is nil")
	}

	installmentNumber, err := strconv.Atoi(strings.ReplaceAll(orderName, "#", ""))
	if err != nil {
		r.logger.Error(err.Error(), zap.String("Name", strings.ReplaceAll(orderName, "#", "")))
		return nil, err
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		r.logger.Error(err.Error(), zap.String("Amount", amountStr))
		return nil, err
	}

	orderID = strings.TrimPrefix(orderID, "gid://shopify/Order/")

	// Pre-create PaymentInstallment
	paymentInstallment := dbModels.PaymentInstallment{
		InstallmentNumber: installmentNumber,
		DueDate:           time.Now().AddDate(0, 1, 0).In(r.loc),
		Amount:            amount,
		Status:            status,
		ShopifyOrderID:    orderID,
		CreatedAt:         time.Now().In(r.loc),
		UpdatedAt:         time.Now().In(r.loc),
	}
	if err := tx.WithContext(ctx).Create(&paymentInstallment).Error; err != nil {
		r.logger.Error("creating payment installment", zap.Error(err), zap.Any("req", paymentInstallment))
		return nil, err
	}

	return &paymentInstallment, nil
}
