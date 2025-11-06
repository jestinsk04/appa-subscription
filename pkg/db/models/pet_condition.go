package models

type PetCondition struct {
	ID        string `gorm:"primaryKey;column:id" json:"id"`
	Name      string `gorm:"column:name" json:"name"`
	PetTypeID string `gorm:"column:pet_type_id" json:"petTypeID"`
}

func (PetCondition) TableName() string {
	return "pets_conditions"
}
