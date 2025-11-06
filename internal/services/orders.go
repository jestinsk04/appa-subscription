package services

import (
	"context"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	dbModels "appa_subscriptions/pkg/db/models"
	"appa_subscriptions/pkg/shopify"
)

type OrderService interface {
	NextPaymentInstallmentCreate(ctx context.Context)
}

type orderService struct {
	db      *gorm.DB
	shopify shopify.Repository
	loc     *time.Location
	logger  *zap.Logger
}

func NewOrderService(db *gorm.DB, shopifyRepo shopify.Repository, loc *time.Location, logger *zap.Logger) OrderService {
	return &orderService{
		db:      db,
		shopify: shopifyRepo,
		loc:     loc,
		logger:  logger,
	}
}

// NextPaymentInstallmentCreate handles the creation of the next payment installment for users policies
func (s *orderService) NextPaymentInstallmentCreate(ctx context.Context) {
	// Implementation for creating the next payment installment
	currentDate := time.Now().In(s.loc)

	var policies []dbModels.Policy
	err := s.db.WithContext(ctx).
		Select("policies.*").
		InnerJoins("User", s.db.Select("ID", "ShopifyID").Model(&dbModels.User{})).
		Where(
			"next_payment_date <= ? AND status = ? AND is_manual = ?",
			currentDate,
			"active",
			true,
		).
		Find(&policies).Error
	if err != nil {
		s.logger.Error(err.Error())
		return
	}

	userPoliciesMap := make(map[string][]dbModels.Policy)
	for _, policy := range policies {
		userPoliciesMap[policy.UserID] = append(userPoliciesMap[policy.UserID], policy)
	}

	var policiesIDs []string
	for userID, policies := range userPoliciesMap {
		email := policies[0].User.Email
		shopifyUserID := policies[0].User.ShopifyID
		lineItems, userPolicyIDs := getOrderLineItemsByPolicies(policies)
		shopifyOrderRequest := shopify.CreateOrderInShopifyRequest{
			Order: shopify.CreateOrderInput{
				AcceptAutomaticDiscounts: false,
				CustomerID:               shopifyUserID,
				Email:                    email,
				Tags:                     []string{"manual_subscription_recurring_order"},
				TaxExempt:                false,
				LineItems:                lineItems,
			},
		}

		order, err := s.shopify.CreateOrder(ctx, shopifyOrderRequest)
		if err != nil {
			s.logger.Error(err.Error(), zap.Any("reques", shopifyOrderRequest))
			continue
		}

		installmentNumber, err := strconv.Atoi(strings.ReplaceAll(order.Name, "#", ""))
		if err != nil {
			s.logger.Error(err.Error(), zap.String("Name", strings.ReplaceAll(order.Name, "#", "")))
			continue
		}

		amount, err := strconv.ParseFloat(order.TotalPriceSet.ShopMoney.Amount, 64)
		if err != nil {
			s.logger.Error(err.Error(), zap.String("Amount", order.TotalPriceSet.ShopMoney.Amount))
			continue
		}

		// create payment installment record in DB
		paymentInstallment := dbModels.PaymentInstallment{
			InstallmentNumber: installmentNumber,
			DueDate:           time.Now().AddDate(0, 1, 0).In(s.loc),
			Amount:            amount,
			Status:            "pending",
			ShopifyOrderID:    strings.TrimPrefix(order.ID, "gid://shopify/Order/"),
			//ShopifyCheckoutURL: order.StatusPageURL, TODO(MARTIN): Ask shopify about this field
			CreatedAt: time.Now().In(s.loc),
			UpdatedAt: time.Now().In(s.loc),
		}
		err = s.db.WithContext(ctx).Create(&paymentInstallment).Error
		if err != nil {
			s.logger.Error(err.Error(), zap.String("user_id", userID))
			continue
		}

		// create policy payment installment records
		policyPayments := []dbModels.PolicyPayment{}
		for _, policy := range policies {
			policyPayments = append(policyPayments, dbModels.PolicyPayment{
				PolicyID:             policy.ID,
				PaymentInstallmentID: paymentInstallment.ID,
				CreatedAt:            time.Now().In(s.loc),
			})
		}

		err = s.db.WithContext(ctx).Create(&policyPayments).Error
		if err != nil {
			s.logger.Error(err.Error())
			continue
		}

		policiesIDs = append(policiesIDs, userPolicyIDs...)
	}

	// update next payment date for policies
	err = s.db.WithContext(ctx).Model(&dbModels.Policy{}).
		Where("id IN ?", policiesIDs).
		Updates(map[string]any{
			"next_payment_date": currentDate.AddDate(0, 1, 0).In(s.loc),
			"updated_at":        time.Now().In(s.loc),
			"status":            "pending_payment",
		}).Error
	if err != nil {
		s.logger.Error(err.Error(), zap.Any("policy_ids", policiesIDs))
		return
	}
}

func getOrderLineItemsByPolicies(policies []dbModels.Policy) ([]shopify.LineItemsNodeRequest, []string) {
	var lineItems []shopify.LineItemsNodeRequest
	var policyIDs []string
	for _, policy := range policies {
		lineItems = append(lineItems, shopify.LineItemsNodeRequest{
			VariantID: policy.ShopifyID,
			Quantity:  1,
		})
		policyIDs = append(policyIDs, policy.ID)
	}
	return lineItems, policyIDs
}
