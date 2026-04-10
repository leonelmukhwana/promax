package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NannyProfile struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	UserID          uuid.UUID      `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	HomeCounty      string         `gorm:"type:text" json:"home_county"` // Changed from bio
	Gender          string         `gorm:"type:varchar(10)" json:"gender"`
	DateOfBirth     time.Time      `json:"dob"`
	Age             int            `json:"age"`
	ExperienceYears int            `json:"experience_years"`
	SelfieURL       string         `json:"selfie_url"`
	IDCardURL       string         `json:"id_card_url"`
	Location        string         `json:"location"`
	Availability    string         `gorm:"default:'available'" json:"availability"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"` // This enables Soft Delete
}

type EmployerProfile struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	UserID         uuid.UUID      `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	IDOrPassportNo string         `json:"id_passport_no"` // Fixed naming
	City           string         `json:"city"`
	Residence      string         `json:"residence"`   // Added residence
	Nationality    string         `json:"nationality"` // Changed to string for country names
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"` // This enables Soft Delete
}
