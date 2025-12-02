package models

import "time"

type User struct {
	ID             string    `gorm:"primaryKey;column:id;default:gen_random_uuid()" json:"id"`
	Name           string    `gorm:"column:name" json:"name"`
	Email          string    `gorm:"column:email" json:"email"`
	Phone          *string   `gorm:"column:phone" json:"phone"`
	City           *string   `gorm:"column:city" json:"city"`
	Role           string    `gorm:"column:role;type:app_role;default:user" json:"role"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
	ShopifyID      string    `gorm:"column:shopify_id" json:"shopifyId"`
	DocumentType   string    `gorm:"column:document_type" json:"documentType"`
	DocumentNumber string    `gorm:"column:document_number" json:"documentNumber"`
	Pets           *[]Pet    `gorm:"foreignKey:UserID;references:ID" json:"pets,omitempty"`
}

func (User) TableName() string {
	return "users"
}
