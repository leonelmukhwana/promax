package models

import (
	"github.com/google/uuid"
	"time"
)

type Complaint struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ReporterID  uuid.UUID `gorm:"type:uuid;not null" json:"reporter_id"` // Who is complaining
	TargetID    uuid.UUID `gorm:"type:uuid;not null" json:"target_id"`   // Who is being complained about
	Category    string    `json:"category"`                              // e.g., "Abuse", "Late Sleeping", "Poor Food", "Rudeness"
	Description string    `gorm:"type:text" json:"description"`
	Status      string    `gorm:"default:'pending'" json:"status"` // pending, investigated, resolved
	CreatedAt   time.Time `json:"created_at"`
}

type Rating struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	EmployerID uuid.UUID `gorm:"type:uuid;not null" json:"employer_id"`
	NannyID    uuid.UUID `gorm:"type:uuid;not null" json:"nanny_id"`
	Score      int       `gorm:"check:score >= 1 AND score <= 5" json:"score"` // 1 to 5 stars
	Comment    string    `json:"comment"`
	Month      string    `json:"month"` // e.g., "April 2026"
	CreatedAt  time.Time `json:"created_at"`
}
