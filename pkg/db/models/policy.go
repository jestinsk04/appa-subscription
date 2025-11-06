package models

import "time"

type Policy struct {
	ID                string    `json:"id" gorm:"column:id;type:uuid;default:gen_random_uuid();not null;primaryKey"`
	UserID            string    `json:"userId" gorm:"column:user_id;type:uuid;not null"`
	PetID             string    `json:"petId" gorm:"column:pet_id;type:uuid;not null"`
	PlanID            string    `json:"planId" gorm:"column:plan_id;type:uuid;not null"`
	StartDate         time.Time `json:"startDate" gorm:"column:start_date;type:date;not null"`
	NextPayment       time.Time `json:"nextPayment" gorm:"column:next_payment;type:date;not null"`
	RemainingBalance  float64   `json:"remainingBalance" gorm:"column:remaining_balance;type:numeric(10,2);not null"`
	Status            string    `json:"status" gorm:"column:status;type:text;default:'active';not null"`
	ShopifyID         string    `json:"shopifyId" gorm:"column:shopify_id;type:text;not null"`
	CreatedAt         time.Time `json:"createdAt" gorm:"column:created_at;type:timestamp with time zone;default:now();not null"`
	UpdatedAt         time.Time `json:"updatedAt" gorm:"column:updated_at;type:timestamp with time zone;default:now();not null"`
	HealthDeclared    bool      `json:"healthDeclared" gorm:"column:health_declared;type:boolean;default:false;not null"`
	LimitPeriodStart  time.Time `json:"limitPeriodStart" gorm:"column:limit_period_start;type:date;default:CURRENT_DATE;not null"`
	LimitPeriodEnd    time.Time `json:"limitPeriodEnd" gorm:"column:limit_period_end;type:date;default:(CURRENT_DATE + interval '1 year');not null"`
	DocumentsVerified bool      `json:"documentsVerified" gorm:"column:documents_verified;type:boolean;default:false;not null"`
	IsManual          bool      `json:"isManual" gorm:"column:is_manual;type:boolean;default:true;not null"`
	User              *User     `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
}

func (Policy) TableName() string {
	return "policies"
}
