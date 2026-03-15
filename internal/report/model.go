package report

import (
	"time"

	"github.com/google/uuid"
)

type Report struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	IssueType    string     `gorm:"size:50;not null" json:"issue_type"`
	Description  string     `json:"description"`
	Latitude     float64    `gorm:"type:decimal(10,8)" json:"latitude"`
	Longitude    float64    `gorm:"type:decimal(11,8)" json:"longitude"`
	Address      string     `json:"address"`
	ImageURL     string     `json:"image_url"`
	Status       string     `gorm:"size:20;default:pending;index" json:"status"`
	PointsEarned int        `gorm:"default:5" json:"points_earned"`
	ResolvedBy   *uuid.UUID `gorm:"type:uuid" json:"resolved_by,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type CreateReportRequest struct {
	IssueType   string  `json:"issue_type" validate:"required"`
	Description string  `json:"description"`
	Latitude    float64 `json:"latitude" validate:"required"`
	Longitude   float64 `json:"longitude" validate:"required"`
	Address     string  `json:"address"`
}

type ReportFilter struct {
	Status    string `form:"status"`
	IssueType string `form:"issue_type"`
	Ward      string `form:"ward"`
	Page      int    `form:"page,default=1"`
	Limit     int    `form:"limit,default=20"`
}

type UpdateStatusRequest struct {
	Status     string `json:"status" validate:"required,oneof=pending assigned resolved"`
	ResolvedBy string `json:"resolved_by,omitempty"`
}
