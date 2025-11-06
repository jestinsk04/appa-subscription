package models

type PetAgeRange struct {
	ID        string `gorm:"primaryKey;column:id;default:gen_random_uuid()" json:"id"`
	Name      string `gorm:"column:name" json:"name"`
	PetTypeID string `gorm:"column:pet_type_id" json:"petTypeID"`
}

func (PetAgeRange) TableName() string {
	return "pets_age_ranges"
}
