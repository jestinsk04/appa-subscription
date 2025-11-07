package models

import "time"

type PaymentInstallment struct {
	ID                 string     `json:"id" gorm:"column:id;default:gen_random_uuid()"`
	InstallmentNumber  int        `json:"installmentNumber" gorm:"column:installment_number"`
	DueDate            time.Time  `json:"dueDate" gorm:"column:due_date"`
	Amount             float64    `json:"amount" gorm:"column:amount"`
	Status             string     `json:"status" gorm:"column:status;default:'pending'"`
	ShopifyOrderID     string     `json:"shopifyOrderId" gorm:"column:shopify_order_id"`
	ShopifyCheckoutURL string     `json:"shopifyCheckoutUrl" gorm:"column:shopify_checkout_url"`
	PaidAt             *time.Time `json:"paidAt,omitempty" gorm:"column:paid_at"`
	CreatedAt          time.Time  `json:"createdAt" gorm:"column:created_at;default:now()"`
	UpdatedAt          time.Time  `json:"updatedAt" gorm:"column:updated_at;default:now()"`
}

func (PaymentInstallment) TableName() string {
	return "payment_installments"
}
