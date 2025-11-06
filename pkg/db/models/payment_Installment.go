package models

import "time"

type PaymentInstallment struct {
	ID                 string    `json:"id" gorm:"column:id;type:uuid;not null;default:gen_random_uuid()"`
	InstallmentNumber  int       `json:"installmentNumber" gorm:"column:installment_number;not null"`
	DueDate            time.Time `json:"dueDate" gorm:"column:due_date;type:date;not null"`
	Amount             float64   `json:"amount" gorm:"column:amount;type:numeric;not null"`
	Status             string    `json:"status" gorm:"column:status;type:text;not null;default:'pending'"`
	ShopifyOrderID     string    `json:"shopifyOrderId,omitempty" gorm:"column:shopify_order_id;type:text"`
	ShopifyCheckoutURL string    `json:"shopifyCheckoutUrl,omitempty" gorm:"column:shopify_checkout_url;type:text"`
	PaidAt             string    `json:"paidAt,omitempty" gorm:"column:paid_at;type:timestamp with time zone"`
	CreatedAt          time.Time `json:"createdAt" gorm:"column:created_at;type:timestamp with time zone;not null;default:now()"`
	UpdatedAt          time.Time `json:"updatedAt" gorm:"column:updated_at;type:timestamp with time zone;not null;default:now()"`
}

func (PaymentInstallment) TableName() string {
	return "payment_installments"
}
