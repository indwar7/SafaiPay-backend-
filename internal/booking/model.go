package booking

import (
	"time"

	"github.com/google/uuid"
)

type Booking struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	CollectorID  *uuid.UUID `gorm:"type:uuid;index" json:"collector_id,omitempty"`
	WasteType    string     `gorm:"size:50;not null" json:"waste_type"`
	BookingDate  time.Time  `gorm:"not null" json:"booking_date"`
	TimeSlot     string     `gorm:"size:30;not null" json:"time_slot"`
	Address      string     `gorm:"not null" json:"address"`
	Latitude     float64    `gorm:"type:decimal(10,8)" json:"latitude"`
	Longitude    float64    `gorm:"type:decimal(11,8)" json:"longitude"`
	Status       string     `gorm:"size:20;default:pending;index" json:"status"`
	Weight       *float64   `gorm:"type:decimal(6,2)" json:"weight,omitempty"`
	PointsEarned int        `gorm:"default:0" json:"points_earned"`
	ImageURL     string     `json:"image_url,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type CreateBookingRequest struct {
	WasteType   string  `json:"waste_type" validate:"required"`
	BookingDate string  `json:"booking_date" validate:"required"`
	TimeSlot    string  `json:"time_slot" validate:"required"`
	Address     string  `json:"address" validate:"required"`
	Latitude    float64 `json:"latitude" validate:"required"`
	Longitude   float64 `json:"longitude" validate:"required"`
}

type UpdateBookingStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending assigned completed cancelled"`
}

type BookingFilter struct {
	Status string `form:"status"`
	Page   int    `form:"page,default=1"`
	Limit  int    `form:"limit,default=20"`
}
