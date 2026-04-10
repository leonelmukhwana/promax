package models

import (
	"github.com/google/uuid"
	"time"
)

type Job struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	EmployerID uuid.UUID `gorm:"type:uuid;not null" json:"employer_id"`
	Employer   User      `gorm:"foreignKey:EmployerID" json:"employer_details,omitempty"` // For Admin view

	NannyID *uuid.UUID `gorm:"type:uuid" json:"nanny_id"`                         // Pointer allows null
	Nanny   User       `gorm:"foreignKey:NannyID" json:"nanny_details,omitempty"` // For Admin view

	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Requirements string    `json:"requirements"`
	Location     string    `json:"location"`
	Salary       float64   `json:"salary"`
	Status       string    `gorm:"default:'open'" json:"status"` // open, assigned, completed
	CreatedAt    time.Time `json:"created_at"`
}

type Booking struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	NannyID    uuid.UUID `gorm:"type:uuid;not null" json:"nanny_id"` // Nanny being vetted
	BookedDate time.Time `gorm:"type:date;not null" json:"booked_date"`
	StartTime  time.Time `gorm:"not null" json:"start_time"`
	EndTime    time.Time `gorm:"not null" json:"end_time"`
	Status     string    `gorm:"default:'pending'" json:"status"` // pending, verified, rejected
	CreatedAt  time.Time `json:"created_at"`
}
