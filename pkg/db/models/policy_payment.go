package models

import "time"

type PolicyPayment struct {
	ID                   string              `gorm:"primaryKey;column:id;default:gen_random_uuid()" json:"id"`
	PolicyID             string              `gorm:"column:policy_id" json:"policyID"`
	PaymentInstallmentID string              `gorm:"column:payment_installment_id" json:"paymentInstallmentID"`
	CreatedAt            time.Time           `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	Policy               *[]Policy           `gorm:"foreignKey:PolicyID;references:ID" json:"policy,omitempty"`
	PaymentInstallment   *PaymentInstallment `gorm:"foreignKey:PaymentInstallmentID;references:ID" json:"paymentInstallment,omitempty"`
}

func (PolicyPayment) TableName() string {
	return "policies_payments"
}
