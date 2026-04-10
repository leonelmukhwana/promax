package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;" json:"id"`
	FirstName  string    `gorm:"not null" json:"first_name"`
	LastName   string    `gorm:"not null" json:"last_name"`
	Email      string    `gorm:"uniqueIndex;not null" json:"email"`
	Phone      string    `gorm:"uniqueIndex;not null" json:"phone"`
	Password   string    `gorm:"not null" json:"-"`
	Role       string    `gorm:"type:varchar(20);not null" json:"role"`
	Status     string    `gorm:"type:varchar(20);default:'inactive'" json:"status"`
	IsVerified bool      `gorm:"default:false" json:"is_verified"`
	// Update this line below:
	CreatedAt time.Time      `gorm:"column:inserted_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// --- RELATIONSHIPS (Add these lines) ---
	// These allow Admin to see the profile data when fetching users
	NannyProfile    *NannyProfile    `gorm:"foreignKey:UserID" json:"nanny_profile,omitempty"`
	EmployerProfile *EmployerProfile `gorm:"foreignKey:UserID" json:"employer_profile,omitempty"`
}

// BeforeCreate hook to generate UUIDs automatically
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.New()
	return
}
