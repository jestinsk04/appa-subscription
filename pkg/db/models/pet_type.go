package models

type PetType struct {
	ID   string `gorm:"primaryKey;column:id;default:gen_random_uuid()" json:"id"`
	Name string `gorm:"column:name" json:"name"`
}

func (PetType) TableName() string {
	return "pets_types"
}
