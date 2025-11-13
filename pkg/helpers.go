package helpers

import (
	"appa_subscriptions/internal/models"
	"fmt"
	"strings"
)

const (
	recurringAppleOrderTagPrefix = "appstle_subscription_recurring_order"
)

// FindRecurringAppleFirstOrderID
func FindRecurringAppleFirstOrderID(tagsStr string) *string {
	tags := strings.SplitSeq(tagsStr, ",")
	for tag := range tags {
		if strings.Contains(tag, recurringAppleOrderTagPrefix) {
			orderTag := strings.TrimPrefix(tag, fmt.Sprintf("%s_", recurringAppleOrderTagPrefix))
			return &orderTag
		}
	}

	return nil
}

func GetVarsForConfirmationOrderEmail(
	varsForEmail models.ConfirmationOrderEmailVars,
) map[string]any {
	vars := make(map[string]any)
	if varsForEmail.FirtsName != "" {
		vars["display_name"] = varsForEmail.FirtsName
	}
	if varsForEmail.PayUrl != "" {
		vars["order_no"] = varsForEmail.PayUrl
	}
	if len(varsForEmail.PetsList) > 0 {
		vars["date"] = varsForEmail.PetsList
	}
	if varsForEmail.DaysLeft > 0 {
		vars["days_left"] = varsForEmail.DaysLeft
	}

	return vars
}
