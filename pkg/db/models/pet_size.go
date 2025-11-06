package models

type PetSize struct {
	ID        string `gorm:"primaryKey;column:id;default:gen_random_uuid()" json:"id"`
	Name      string `gorm:"column:name" json:"name"`
	PetTypeID string `gorm:"column:pet_type_id" json:"petTypeID"`
}

func (PetSize) TableName() string {
	return "pets_sizes"
}
