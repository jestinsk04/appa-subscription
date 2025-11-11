package helpers

import (
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
