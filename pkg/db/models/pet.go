package models

import "time"

type Pet struct {
	ID          string        `gorm:"primaryKey;column:id;default:gen_random_uuid()" json:"id"`
	UserID      string        `gorm:"column:user_id" json:"userID"`
	Name        string        `gorm:"column:name" json:"name"`
	Breed       string        `gorm:"column:breed" json:"breed"`
	Gender      string        `gorm:"column:gender" json:"gender"`
	Weight      float64       `gorm:"column:weight" json:"weight"`
	MicrochipID string        `gorm:"column:microchip_id" json:"microchipID"`
	AgeRangeID  string        `gorm:"column:age_range_id" json:"ageRangeID"`
	ConditionID string        `gorm:"column:condition_id" json:"conditionID"`
	SizeID      string        `gorm:"column:size_id" json:"sizeID"`
	TypeID      string        `gorm:"column:type_id" json:"typeID"`
	CreatedAt   time.Time     `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt   time.Time     `gorm:"column:updated_at" json:"updatedAt"`
	AgeRange    *PetAgeRange  `gorm:"foreignKey:AgeRangeID;references:ID" json:"ageRange,omitempty"`
	Condition   *PetCondition `gorm:"foreignKey:ConditionID;references:ID" json:"condition,omitempty"`
	Size        *PetSize      `gorm:"foreignKey:SizeID;references:ID" json:"size,omitempty"`
	User        *User         `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	Type        *PetType      `gorm:"foreignKey:TypeID;references:ID" json:"type,omitempty"`
}

func (Pet) TableName() string {
	return "pets"
}
