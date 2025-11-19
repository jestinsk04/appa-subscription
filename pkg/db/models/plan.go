package models

type Plan struct {
	ID           string  `gorm:"primaryKey;column:id;default:gen_random_uuid()" json:"id"`
	Name         string  `gorm:"column:name" json:"name"`
	MonthlyPrice float64 `gorm:"column:monthly_price" json:"monthlyPrice"`
	AnnualLimit  float64 `gorm:"column:annual_limit" json:"annualLimit"`
	Description  string  `gorm:"column:description" json:"description"`
	ShopifyID    string  `gorm:"column:shopify_id" json:"shopifyID"`
	PetTypeID    string  `gorm:"column:pet_type_id" json:"petTypeID"`
	CreatedAt    string  `gorm:"column:created_at" json:"createdAt"`
}

func (Plan) TableName() string {
	return "plans"
}
