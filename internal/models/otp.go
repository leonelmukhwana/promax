package models

import (
	"github.com/google/uuid"
	"time"
)

type OTP struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"`
	Code      string    `gorm:"not null"`
	Purpose   string    `gorm:"not null"` // "verification" or "reset"
	ExpiresAt time.Time `gorm:"not null"`
}
