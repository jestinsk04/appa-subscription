package services

import (
	"context"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"appa_subscriptions/internal/domains"
	"appa_subscriptions/pkg/db"
	dbModels "appa_subscriptions/pkg/db/models"
	PaymentInstallment "appa_subscriptions/pkg/db/repositories"
	"appa_subscriptions/pkg/shopify"
)

type orderService struct {
	db                     *gorm.DB
	loc                    *time.Location
	shopify                shopify.Repository
	PaymentInstallmentRepo PaymentInstallment.Repository
	logger                 *zap.Logger
}

func NewOrderService(
	db *gorm.DB,
	shopifyRepo shopify.Repository,
	PaymentInstallmentRepo PaymentInstallment.Repository,
	loc *time.Location,
	logger *zap.Logger,
) domains.OrderService {
	return &orderService{
		db:                     db,
		shopify:                shopifyRepo,
		PaymentInstallmentRepo: PaymentInstallmentRepo,
		loc:                    loc,
		logger:                 logger,
	}
}

// NextPaymentInstallmentCreate handles the creation of the next payment installment for users policies
func (s *orderService) NextPaymentInstallmentCreate(ctx context.Context) error {
	currentDate := time.Now().In(s.loc)

	var policies []dbModels.Policy
	err := s.db.WithContext(ctx).
		Select("policies.*").
		InnerJoins("User", s.db.Select("ID", "ShopifyID").Model(&dbModels.User{})).
		Where(
			"next_payment <= ? AND status = ? AND is_manual = ?",
			currentDate,
			"active",
			true,
		).
		Find(&policies).Error
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}

	userPoliciesMap := make(map[string][]dbModels.Policy)
	for _, policy := range policies {
		userPoliciesMap[policy.UserID] = append(userPoliciesMap[policy.UserID], policy)
	}

	for _, policies := range userPoliciesMap {
		var (
			email         = policies[0].User.Email
			shopifyUserID = policies[0].User.ShopifyID
		)

		order, userPolicyIDs, err := s.createOrderInShopify(
			ctx,
			policies,
			email,
			shopifyUserID,
		)
		if err != nil {
			s.logger.Error(err.Error(), zap.Any("policies", policies))
			continue
		}

		var (
			tx    = s.db.Begin().WithContext(ctx)
			errDB error
		)
		// defer db.DBRollback(tx, &errDB)

		paymentInstallment, err := s.PaymentInstallmentRepo.Create(
			tx, ctx, order.ID, statusPendingPayment, order.Name, order.TotalPriceSet.ShopMoney.Amount,
		)
		if err != nil {
			errDB = err
			s.logger.Error(err.Error(), zap.Any("order", order))
			continue
		}

		// create policy payment installment records
		policyPayments := s.getPolicyPaymentsByPolicies(policies, paymentInstallment.ID)
		err = tx.WithContext(ctx).Create(&policyPayments).Error
		if err != nil {
			s.logger.Error(err.Error())
			continue
		}

		// update next payment date for policies
		err = tx.WithContext(ctx).Model(&dbModels.Policy{}).
			Where("id IN ?", userPolicyIDs).
			Updates(map[string]any{
				"next_payment": currentDate.AddDate(0, 1, 0).In(s.loc).Format("2006-01-02"),
				"status":       statusPendingPolicy,
			}).Error
		if err != nil {
			s.logger.Error(err.Error(), zap.Any("policy_ids", userPolicyIDs))
			return err
		}

		db.DBRollback(tx, &errDB)
	}

	return nil
}

// getPolicyPaymentsByPolicies creates PolicyPayment records for the given policies and payment installment ID
func (s *orderService) getPolicyPaymentsByPolicies(
	policies []dbModels.Policy,
	paymentInstallmentID string,
) []dbModels.PolicyPayment {
	var policyPayments []dbModels.PolicyPayment
	for _, policy := range policies {
		policyPayments = append(policyPayments, dbModels.PolicyPayment{
			PolicyID:             policy.ID,
			PaymentInstallmentID: paymentInstallmentID,
		})
	}
	return policyPayments
}

// createOrderInShopify creates an order in Shopify for the given user and line items
func (s *orderService) createOrderInShopify(
	ctx context.Context,
	policies []dbModels.Policy,
	email,
	shopifyUserID string,
) (*shopify.OrderCreateResponse, []string, error) {
	lineItems, userPolicyIDs := getOrderLineItemsByPolicies(policies)

	if !strings.Contains(shopifyUserID, shopify.CustomerKind) {
		shopifyUserID = shopify.GID(shopify.CustomerKind, shopifyUserID)
	}

	shopifyOrderRequest := shopify.CreateOrderInShopifyRequest{
		Order: shopify.CreateOrderInput{
			CustomerID:      shopifyUserID,
			Email:           email,
			Tags:            []string{tagManualSubscriptionRecurringOrder},
			LineItems:       lineItems,
			Note:            "Order created for manual subscription recurring payment",
			FinancialStatus: "PENDING",
		},
	}

	order, err := s.shopify.CreateOrder(ctx, shopifyOrderRequest)
	if err != nil {
		s.logger.Error(err.Error(), zap.Any("reques", shopifyOrderRequest))
		return nil, nil, err
	}

	return order, userPolicyIDs, nil
}

func getOrderLineItemsByPolicies(policies []dbModels.Policy) ([]shopify.LineItemsNodeRequest, []string) {
	var lineItems []shopify.LineItemsNodeRequest
	var policyIDs []string
	for _, policy := range policies {
		if !strings.Contains(policy.ShopifyID, shopify.ProductVariantKind) {
			policy.ShopifyID = shopify.GID(shopify.ProductVariantKind, policy.ShopifyID)
		}

		lineItems = append(lineItems, shopify.LineItemsNodeRequest{
			VariantID: policy.ShopifyID,
			Quantity:  1,
		})
		policyIDs = append(policyIDs, policy.ID)
	}
	return lineItems, policyIDs
}
