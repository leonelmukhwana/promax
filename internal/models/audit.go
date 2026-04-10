package models

import (
	"github.com/google/uuid"
	"time"
)

type AuditLog struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;index"`
	Action    string    `gorm:"type:varchar(100)"` // e.g., "USER_REGISTER", "PASSWORD_RESET"
	Resource  string    `gorm:"type:varchar(100)"` // e.g., "Users", "Jobs"
	OldValue  string    `gorm:"type:text"`         // Store as JSON string
	NewValue  string    `gorm:"type:text"`         // Store as JSON string
	IPAddress string    `gorm:"type:varchar(45)"`
	CreatedAt time.Time `gorm:"index"`
}
