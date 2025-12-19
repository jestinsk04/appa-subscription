package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"appa_subscriptions/internal/domains"
	"appa_subscriptions/internal/models"
	helpers "appa_subscriptions/pkg"
	"appa_subscriptions/pkg/db"
	dbModels "appa_subscriptions/pkg/db/models"
	PaymentInstallment "appa_subscriptions/pkg/db/repositories"
	"appa_subscriptions/pkg/mailgun"
	"appa_subscriptions/pkg/shopify"
)

type orderService struct {
	db                     *gorm.DB
	shopify                shopify.Repository
	PaymentInstallmentRepo PaymentInstallment.Repository
	muRepo                 mailgun.Repository
	loc                    *time.Location
	logger                 *zap.Logger
}

const (
	statusCanceled = "canceled"
)

func NewOrderService(
	db *gorm.DB,
	shopifyRepo shopify.Repository,
	PaymentInstallmentRepo PaymentInstallment.Repository,
	mailgunRepo mailgun.Repository,
	loc *time.Location,
	logger *zap.Logger,
) domains.OrderService {
	return &orderService{
		db:                     db,
		shopify:                shopifyRepo,
		PaymentInstallmentRepo: PaymentInstallmentRepo,
		muRepo:                 mailgunRepo,
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
		InnerJoins("Pet", s.db.Select("Name").Model(&dbModels.Pet{})).
		InnerJoins("User", s.db.Select("ID", "ShopifyID", "Email").Model(&dbModels.User{})).
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
			errDB error
			tx    = s.db.Begin().WithContext(ctx)
		)

		paymentInstallment, err := s.PaymentInstallmentRepo.Create(
			tx, ctx, order.ID, statusPendingPayment, order.Name, order.TotalPriceSet.ShopMoney.Amount,
		)
		if err != nil {
			errDB = err
			db.DBRollback(tx, &errDB)
			s.logger.Error(err.Error(), zap.Any("order", order))
			continue
		}

		// create policy payment installment records
		policyPayments := s.getPolicyPaymentsByPolicies(policies, paymentInstallment.ID)
		err = tx.WithContext(ctx).Create(&policyPayments).Error
		if err != nil {
			errDB = err
			db.DBRollback(tx, &errDB)
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
			errDB = err
			db.DBRollback(tx, &errDB)
			s.logger.Error(err.Error(), zap.Any("policy_ids", userPolicyIDs))
			return err
		}

		db.DBRollback(tx, &errDB)

		pets := make([]string, len(policies))
		for _, policy := range policies {
			pets = append(pets, policy.Pet.Name)
		}

		notificationJobsQueue <- notificationJob{
			vars: models.ConfirmationOrderEmailVars{
				FirtsName: policies[0].User.Name,
				PetsList:  pets,
				PayUrl: fmt.Sprintf(
					"https://pay.appasalud.com/?orderId=%s",
					strings.TrimPrefix(order.ID, "gid://shopify/Order/"),
				),
			},
			email:    email,
			template: "create_order",
		}
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

func (s *orderService) ReminderPendingPolicies(ctx context.Context) error {
	s.logger.Info("starting ReminderPendingPolicies process")
	now := time.Now().In(s.loc)

	var policies []dbModels.PolicyPayment
	err := s.db.WithContext(ctx).Debug().
		Select("policies_payments.*").
		InnerJoins("PaymentInstallment", s.db.Select("ShopifyOrderID").Model(&dbModels.PaymentInstallment{})).
		InnerJoins("Policy", s.db.Select("ID").Where(&dbModels.Policy{
			Status:   statusPendingPolicy,
			IsManual: true,
		})).
		Joins("Policy.User", s.db.Select("Email", "Name").Model(&dbModels.User{})).
		Joins("Policy.Pet", s.db.Select("Name").Model(&dbModels.Pet{})).
		Find(&policies).Error
	if err != nil {
		s.logger.Error("failed to fetch policies by next_delivery_day", zap.Error(err))
		return err
	}

	policyPetsMap := make(map[string][]string)
	for _, policyPayment := range policies {
		policyPetsMap[policyPayment.PaymentInstallmentID] = append(
			policyPetsMap[policyPayment.PaymentInstallmentID],
			policyPayment.Policy.Pet.Name,
		)
	}

	result := make(map[string]bool)
	for _, policyPayment := range policies {

		if _, exist := result[policyPayment.PaymentInstallmentID]; exist {
			continue
		}

		var (
			nextPaymentDay = policyPayment.Policy.NextPayment.In(s.loc)
			daysPending    = int(now.Sub(nextPaymentDay).Hours() / 24)
			template       string
		)

		switch {
		case daysPending >= 1 && daysPending <= 27:
			template = "reminder"
		case daysPending == 31:
			template = "cancellation"
			daysPending = 30
			err := s.UpdatePoliceStatus(ctx, policyPayment.Policy.ID, statusCanceled)
			if err != nil {
				s.logger.Error("failed to update policy status to canceled", zap.String("policy_id", policyPayment.Policy.ID))
			}
		case daysPending > 31:
			daysPending = 0
			template = "reactivation"
		default:
			continue
		}

		result[policyPayment.PaymentInstallmentID] = true
		notificationJobsQueue <- notificationJob{
			vars: models.ConfirmationOrderEmailVars{
				FirtsName: policyPayment.Policy.User.Name,
				DaysLeft:  30 - daysPending,
				PetsList:  policyPetsMap[policyPayment.PaymentInstallmentID],
				PayUrl: fmt.Sprintf(
					"https://pay.appasalud.com/?orderId=%s",
					policyPayment.PaymentInstallment.ShopifyOrderID,
				),
			},
			email:    policyPayment.Policy.User.Email,
			template: template,
		}
	}

	s.logger.Info("completed ReminderPendingPolicies process")
	return nil
}

// UpdatePoliceStatus updates the status of a policy
func (o *orderService) UpdatePoliceStatus(
	ctx context.Context,
	policyID string,
	status string,
) error {
	err := o.db.WithContext(ctx).Model(&dbModels.Policy{}).
		Where("id = ?", policyID).
		Update("status", status).Error
	if err != nil {
		o.logger.Error(err.Error(), zap.String("policy_id", policyID))
		return err
	}

	return nil
}

// SendEmail sends an email to the customer
func (o *orderService) SendEmail(
	ctx context.Context,
	vars models.ConfirmationOrderEmailVars,
	email string,
	template string,
) error {
	varsForEmail := helpers.GetVarsForConfirmationOrderEmail(vars)

	emailTemplate := EmailsTemplates[template]
	emailTemplate.Vars = varsForEmail
	emailTemplate.To = email

	err := o.muRepo.SendEmail(ctx, emailTemplate)
	if err != nil {
		//o.logger.Error(err.Error(), zap.String("to", order.Order.Customer.Email))
		return err
	}

	return nil
}
