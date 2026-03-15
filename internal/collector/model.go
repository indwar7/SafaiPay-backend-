package collector

import (
	"time"

	"github.com/google/uuid"
)

type Collector struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PhoneNumber       string    `gorm:"uniqueIndex;size:15;not null" json:"phone_number"`
	Name              string    `gorm:"size:100" json:"name"`
	Ward              string    `gorm:"size:50;index" json:"ward"`
	CurrentLat        float64   `gorm:"type:decimal(10,8)" json:"current_lat"`
	CurrentLng        float64   `gorm:"type:decimal(11,8)" json:"current_lng"`
	Status            string    `gorm:"size:20;default:available;index" json:"status"`
	Rating            float64   `gorm:"type:decimal(3,2);default:5.0" json:"rating"`
	TotalCollected    float64   `gorm:"type:decimal(10,2);default:0" json:"total_collected"`
	FCMToken          string    `json:"-"`
	BankAccountNumber string    `gorm:"size:20" json:"-"`
	BankIFSC          string    `gorm:"size:15" json:"-"`
	BankName          string    `gorm:"size:100" json:"-"`
	IsVerified        bool      `gorm:"default:false" json:"is_verified"`
	CreatedAt         time.Time `json:"created_at"`
}

type CollectorResponse struct {
	ID             uuid.UUID `json:"id"`
	PhoneNumber    string    `json:"phone_number"`
	Name           string    `json:"name"`
	Ward           string    `json:"ward"`
	Status         string    `json:"status"`
	Rating         float64   `json:"rating"`
	TotalCollected float64   `json:"total_collected"`
	IsVerified     bool      `json:"is_verified"`
	CreatedAt      time.Time `json:"created_at"`
}

type CompleteBookingRequest struct {
	Weight float64 `json:"weight" validate:"required,min=0.1"`
	Notes  string  `json:"notes"`
}

type UpdateLocationRequest struct {
	Latitude  float64 `json:"latitude" validate:"required"`
	Longitude float64 `json:"longitude" validate:"required"`
}

func (c *Collector) ToResponse() *CollectorResponse {
	return &CollectorResponse{
		ID:             c.ID,
		PhoneNumber:    c.PhoneNumber,
		Name:           c.Name,
		Ward:           c.Ward,
		Status:         c.Status,
		Rating:         c.Rating,
		TotalCollected: c.TotalCollected,
		IsVerified:     c.IsVerified,
		CreatedAt:      c.CreatedAt,
	}
}
