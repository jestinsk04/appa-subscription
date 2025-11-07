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
	"appa_subscriptions/pkg/db"
	dbModels "appa_subscriptions/pkg/db/models"
	PaymentInstallment "appa_subscriptions/pkg/db/repositories"
	"appa_subscriptions/pkg/shopify"
)

const (
	statusPendingPolicy                  = "payment_pending"
	statusPendingPayment                 = "pending"
	statusPaidPayment                    = "paid"
	statusActive                         = "active"
	tagManualSubscriptionRecurringOrder  = "manual_subscription_recurring_order"
	tagAppstleSubscriptionRecurringOrder = "appstle_subscription_recurring_order"
)

type webhookService struct {
	db                     *gorm.DB
	loc                    *time.Location
	ShopifyRepository      shopify.Repository
	PaymentInstallmentRepo PaymentInstallment.Repository
	logger                 *zap.Logger
}

func NewWebhookService(
	db *gorm.DB,
	loc *time.Location,
	shopifyRepo shopify.Repository,
	PaymentInstallmentRepo PaymentInstallment.Repository,
	logger *zap.Logger,
) domains.WebhookService {
	return &webhookService{
		db:                     db,
		loc:                    loc,
		ShopifyRepository:      shopifyRepo,
		PaymentInstallmentRepo: PaymentInstallmentRepo,
		logger:                 logger,
	}
}

// OrderCreated handles the order created webhook from Shopify.
func (s *webhookService) OrderCreated(webhook models.Webhook) {
	s.logger.Info("received order created webhook", zap.Int("order_id", webhook.ID))
	var (
		ctx           = context.Background()
		paymentStatus = statusPendingPayment
		policyStatus  = statusPendingPolicy
	)

	if strings.ToUpper(webhook.FinancialStatus) == "PAID" {
		paymentStatus = statusPaidPayment
		policyStatus = statusActive
	}

	tags := strings.ToLower(webhook.Tags)
	if strings.Contains(tags, tagManualSubscriptionRecurringOrder) {
		return // Skip processing for recurring orders
	}
	if strings.Contains(tags, tagAppstleSubscriptionRecurringOrder) {
		s.orderRecurring(ctx, webhook, paymentStatus, policyStatus)
		return
	}
	s.firstOrderProcess(ctx, webhook, paymentStatus, policyStatus)

	s.logger.Info("completed processing order created webhook", zap.Int("order_id", webhook.ID))
}

// OrderPaid handles the order paid webhook from Shopify.
func (s *webhookService) OrderPaid(
	webhook models.Webhook,
) {
	webhook.ID = 6563075358970
	// 1. Get PolicyPayments by Shopify Order ID
	var policyPayments []dbModels.PolicyPayment
	err := s.db.WithContext(context.Background()).
		Select("policies_payments.*").
		InnerJoins("PaymentInstallment", s.db.Select("ID").Where(&dbModels.PaymentInstallment{
			ShopifyOrderID: fmt.Sprintf("%d", webhook.ID),
		})).
		Find(&policyPayments).Error
	if err != nil {
		s.logger.Error("getting policy payments by shopify order ID", zap.Error(err))
		return
	}

	if len(policyPayments) == 0 {
		s.logger.Warn("no policy payments found for shopify order ID", zap.String("shopify_order_id", fmt.Sprintf("%d", webhook.ID)))
		return
	}

	// var (
	// 	tx    = s.db.Begin().WithContext(context.Background())
	// 	errDB error
	// )
	// defer db.DBRollback(tx, &errDB)

	// 2. Update PaymentInstallment status to 'paid'
	err = s.db.WithContext(context.Background()).Model(&dbModels.PaymentInstallment{}).
		Where("id = ?", policyPayments[0].PaymentInstallmentID).
		Updates(map[string]any{
			"status":  statusPaidPayment,
			"paid_at": time.Now().In(s.loc),
		}).Error
	if err != nil {
		s.logger.Error("updating payment installment status to paid", zap.Error(err))
		return
	}

	// // 3. Update Policies status to 'active'
	// policiesIDs := make([]string, 0)
	// for _, pp := range policyPayments {
	// 	policiesIDs = append(policiesIDs, pp.PolicyID)
	// }

	// errDB = tx.Model(&dbModels.Policy{}).
	// 	Where("id IN ?", policiesIDs).
	// 	Updates(map[string]any{
	// 		"status": statusActive,
	// 	}).Error
	// if errDB != nil {
	// 	s.logger.Error("updating policies status to active", zap.Error(errDB))
	// 	return
	// }
}

// firstOrderProcess handles the first order process webhook from Shopify.
func (s *webhookService) firstOrderProcess(
	ctx context.Context,
	webhook models.Webhook,
	paymentStatus, policyStatus string,
) {
	var count int64
	err := s.db.WithContext(ctx).Model(&dbModels.PaymentInstallment{}).
		Where("shopify_order_id = ?", fmt.Sprintf("%d", webhook.ID)).
		Count(&count).Error
	if err != nil {
		s.logger.Error("checking existing payment installment", zap.Error(err))
		return
	}
	if count > 0 {
		s.logger.Info("payment installment already exists for this order, skipping", zap.Int("order_id", webhook.ID))
		return
	}

	var (
		errDB    error
		isManual bool
		tx       = s.db.Begin().WithContext(ctx)
	)
	defer db.DBRollback(tx, &errDB)

	if strings.Contains(strings.ToLower(webhook.Tags), "appstle_subscription") {
		isManual = true
	}

	// 1. Get Dog Data from Shopify Metafields
	pets, err := s.ShopifyRepository.GetDogData(ctx, fmt.Sprintf("%d", webhook.ID))
	if err != nil {
		errDB = err
		return
	}

	// 2. Create or get user
	user, err := s.createOrFindUser(
		tx, ctx, webhook.Customer,
	)
	if err != nil {
		errDB = err
		return
	}

	// 3. Precreate PaymentInstallment
	paymentInstallment, err := s.PaymentInstallmentRepo.Create(
		tx, ctx, fmt.Sprintf("%d", webhook.ID), paymentStatus, webhook.Name, webhook.CurrentTotalPriceSet.ShopMoney.Amount,
	)
	if err != nil {
		errDB = err
		return
	}

	// 4. Get Variants map by Line Items
	variantsMap, err := s.getVariantsMapByLineItem(
		ctx, webhook.LineItems,
	)
	if err != nil {
		errDB = err
		return
	}

	// 5. Register Pets policies
	policyPayments := make([]dbModels.PolicyPayment, 0)
	for _, pet := range pets.Pets {
		petAttributesMap, err := s.GetPetAttributesIDsByVariantID(
			ctx, variantsMap, pet.ProductVariantID, pet.Type,
		)
		if err != nil {
			continue
		}

		dbPet, err := s.createOrFindPet(
			tx, ctx, pet, user.ID,
			petAttributesMap["age"],
			petAttributesMap["size"],
			petAttributesMap["condition"],
			petAttributesMap["type"],
		)
		if err != nil {
			errDB = err
			return
		}

		policy, err := s.createPolicy(
			tx, ctx, isManual, policyStatus, user.ID, dbPet.ID,
			petAttributesMap["plan"], pet.ProductVariantID,
		)
		if err != nil {
			errDB = err
			return
		}

		policyPayments = append(policyPayments, dbModels.PolicyPayment{
			PolicyID:             policy.ID,
			PaymentInstallmentID: paymentInstallment.ID,
			CreatedAt:            time.Now().In(s.loc),
		})
	}

	// 6. Create PolicyPayments
	errDB = tx.WithContext(ctx).Create(&policyPayments).Error
	if errDB != nil {
		s.logger.Error("creating policy payments", zap.Error(errDB))
	}
}

// orderRecurring handles the order recurring webhook from Shopify.
func (s *webhookService) orderRecurring(
	ctx context.Context,
	webhook models.Webhook,
	paymentStatus, policyStatus string,
) {
	// 1. Get first order ID from tags
	tags := strings.Split(webhook.Tags, "-")
	if len(tags) < 2 {
		s.logger.Error("invalid tags format for recurring order", zap.String("tags", webhook.Tags))
		return
	}

	// 2. Get PolicyPayments by first Shopify Order ID
	firstOrderID := tags[1]
	var policyPayments []dbModels.PolicyPayment
	err := s.db.WithContext(context.Background()).
		Select("policy_payments.*").
		InnerJoins("PaymentInstallment").
		Where("payment_installments.shopify_order_id = ?", firstOrderID).
		Find(&policyPayments).Error
	if err != nil {
		s.logger.Error("getting policy payments by shopify order ID", zap.Error(err))
		return
	}

	if len(policyPayments) == 0 {
		s.logger.Warn("no policy payments found for shopify order ID", zap.String("shopify_order_id", firstOrderID))
		return
	}

	var (
		tx    = s.db.Begin().WithContext(context.Background())
		errDB error
	)
	defer db.DBRollback(tx, &errDB)

	// 3. Precreate PaymentInstallment
	paymentInstallment, err := s.PaymentInstallmentRepo.Create(
		tx, ctx, fmt.Sprintf("%d", webhook.ID), paymentStatus, webhook.Name, webhook.CurrentTotalPriceSet.ShopMoney.Amount,
	)
	if err != nil {
		errDB = err
		return
	}

	// 4. Update Policies status
	policiesIDs := make([]string, 0)
	newPolicyPayments := make([]dbModels.PolicyPayment, 0)
	for _, pp := range policyPayments {
		policiesIDs = append(policiesIDs, pp.PolicyID)
		newPolicyPayments = append(newPolicyPayments, dbModels.PolicyPayment{
			PolicyID:             pp.ID,
			PaymentInstallmentID: paymentInstallment.ID,
			CreatedAt:            time.Now().In(s.loc),
		})
	}

	// 5. Update Policies status
	errDB = tx.Model(&dbModels.Policy{}).
		Where("id IN ?", policiesIDs).
		Updates(map[string]any{
			"status": policyStatus,
		}).Error
	if errDB != nil {
		s.logger.Error("updating policies status to active", zap.Error(errDB))
		return
	}

	errDB = tx.WithContext(ctx).Create(&newPolicyPayments).Error
	if errDB != nil {
		s.logger.Error("creating policy payments", zap.Error(errDB))
	}
}

// createOrFindUser creates a new user or finds an existing one based on the Shopify customer ID.
func (s *webhookService) createOrFindUser(
	tx *gorm.DB, ctx context.Context, customer models.Customer,
) (*dbModels.User, error) {
	user := dbModels.User{
		Name:      fmt.Sprintf("%s %s", customer.FirstName, customer.LastName),
		Email:     customer.Email,
		Phone:     &customer.DefaultAddress.Phone,
		City:      &customer.DefaultAddress.City,
		CreatedAt: time.Now().In(s.loc),
		UpdatedAt: time.Now().In(s.loc),
		ShopifyID: fmt.Sprintf("%d", customer.ID),
		Pets:      nil,
	}
	err := tx.WithContext(ctx).Where(dbModels.User{ShopifyID: user.ShopifyID}).FirstOrCreate(&user).Error
	if err != nil {
		s.logger.Error("creating user", zap.Error(err))
		return nil, err
	}

	return &user, nil
}

// getVariantsMapByLineItem
func (s *webhookService) getVariantsMapByLineItem(
	ctx context.Context, lineItems []models.LineItem,
) (map[string]models.Variant, error) {
	variantsMap := make(map[string]models.Variant)
	for _, item := range lineItems {
		if _, exist := variantsMap[fmt.Sprintf("%d", item.VariantID)]; exist {
			continue
		}

		variant, err := s.ShopifyRepository.GetVariantByID(ctx, fmt.Sprintf("%d", item.VariantID))
		if err != nil {
			s.logger.Error("getting variant by ID", zap.Error(err))
			return nil, err
		}

		variantsMap[fmt.Sprintf("%d", item.VariantID)] = models.Variant{
			ID:           variant.ID,
			ProductID:    fmt.Sprintf("%d", item.ProductID),
			PetAge:       shopify.GetDogDataAgeOption(variant.SelectedOptions),
			PetSize:      shopify.GetDogDataSizeOption(variant.SelectedOptions),
			PetCondition: shopify.GetDogDataConditionOption(variant.SelectedOptions),
		}
	}

	return variantsMap, nil
}

// GetPetAttributesIDsByVariantID
func (s *webhookService) GetPetAttributesIDsByVariantID(
	ctx context.Context,
	variantsMap map[string]models.Variant,
	variantID string,
	typeStr string,
) (map[string]string, error) {
	var (
		petAgeRange   dbModels.PetAgeRange
		petSize       dbModels.PetSize
		petCondition  dbModels.PetCondition
		petType       dbModels.PetType
		plan          dbModels.Plan
		attributesMap = make(map[string]string)
	)

	variant, exist := variantsMap[variantID]
	if !exist {
		s.logger.Error("variant data not found for pet", zap.String("variant_id", variantID))
		return nil, fmt.Errorf("variant data not found for variant ID: %s", variantID)
	}

	err := s.db.WithContext(ctx).
		Where(dbModels.PetType{Name: strings.ToLower(typeStr)}).First(&petType).Error
	if err != nil {
		s.logger.Error("getting pet type", zap.Error(err))
		return nil, err
	}

	err = s.db.WithContext(ctx).
		Where(dbModels.PetAgeRange{Name: strings.ToLower(variant.PetAge)}).First(&petAgeRange).Error
	if err != nil {
		s.logger.Error("getting pet age range", zap.Error(err))
		return nil, err
	}

	err = s.db.WithContext(ctx).
		Where(dbModels.PetSize{Name: strings.ToLower(variant.PetSize)}).First(&petSize).Error
	if err != nil {
		s.logger.Error("getting pet size", zap.Error(err))
		return nil, err
	}

	err = s.db.WithContext(ctx).
		Where(dbModels.PetCondition{Name: strings.ToLower(variant.PetCondition)}).First(&petCondition).Error
	if err != nil {
		s.logger.Error("getting pet condition", zap.Error(err))
		return nil, err
	}

	err = s.db.WithContext(ctx).
		Where(dbModels.Plan{ShopifyID: variant.ProductID}).First(&plan).Error
	if err != nil {
		s.logger.Error("getting plan", zap.Error(err))
		return nil, err
	}

	attributesMap["age"] = petAgeRange.ID
	attributesMap["size"] = petSize.ID
	attributesMap["condition"] = petCondition.ID
	attributesMap["plan"] = plan.ID
	attributesMap["type"] = petType.ID

	return attributesMap, nil
}

// createOrFindPet
func (s *webhookService) createOrFindPet(
	tx *gorm.DB,
	ctx context.Context,
	pet shopify.Pet,
	userID, ageRangeID, sizeID, conditionID, typeID string,
) (*dbModels.Pet, error) {
	dbPet := dbModels.Pet{
		Name:        pet.Name,
		Breed:       pet.Breed,
		Gender:      pet.Gender,
		CreatedAt:   time.Now().In(s.loc),
		AgeRangeID:  ageRangeID,
		SizeID:      sizeID,
		ConditionID: conditionID,
		TypeID:      typeID,
		UserID:      userID,
	}
	if err := tx.WithContext(ctx).
		Where(dbModels.Pet{Name: dbPet.Name, UserID: userID}).
		Omit("MicrochipID").
		FirstOrCreate(&dbPet).Error; err != nil {
		s.logger.Error("creating pet", zap.Error(err))
		return nil, err
	}

	return &dbPet, nil
}

// createPolicy
func (s *webhookService) createPolicy(
	tx *gorm.DB,
	ctx context.Context,
	isManual bool,
	status, userID, petID, planID, variantID string,
) (*dbModels.Policy, error) {
	policy := dbModels.Policy{
		UserID:           userID,
		PetID:            petID,
		PlanID:           planID,
		StartDate:        time.Now().In(s.loc),
		NextPayment:      time.Now().AddDate(0, 1, 0).In(s.loc),
		Status:           status,
		HealthDeclared:   false,
		ShopifyID:        variantID,
		LimitPeriodStart: time.Now().In(s.loc),
		LimitPeriodEnd:   time.Now().AddDate(1, 0, 0).In(s.loc),
		CreatedAt:        time.Now().In(s.loc),
		IsManual:         isManual,
	}
	err := tx.WithContext(ctx).Create(&policy).Error
	if err != nil {
		s.logger.Error("creating policy", zap.Error(err))
		return nil, err
	}

	return &policy, nil
}
