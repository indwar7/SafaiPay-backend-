package user

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	PhoneNumber   string         `gorm:"uniqueIndex;size:15;not null" json:"phone_number"`
	Name          string         `gorm:"size:100" json:"name"`
	Ward          string         `gorm:"size:50;index" json:"ward"`
	Address       string         `json:"address"`
	Points        int            `gorm:"default:0" json:"points"`
	WalletBalance float64        `gorm:"type:decimal(10,2);default:0" json:"wallet_balance"`
	TotalReports  int            `gorm:"default:0" json:"total_reports"`
	TotalBookings int            `gorm:"default:0" json:"total_bookings"`
	Streak        int            `gorm:"default:0" json:"streak"`
	LastCheckIn   *time.Time     `json:"last_check_in"`
	FCMToken      string         `json:"-"`
	IsActive      bool           `gorm:"default:true" json:"is_active"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

type UpdateProfileRequest struct {
	Name     *string `json:"name"`
	Ward     *string `json:"ward"`
	Address  *string `json:"address"`
	FCMToken *string `json:"fcm_token"`
}

type UserResponse struct {
	ID            uuid.UUID  `json:"id"`
	PhoneNumber   string     `json:"phone_number"`
	Name          string     `json:"name"`
	Ward          string     `json:"ward"`
	Address       string     `json:"address"`
	Points        int        `json:"points"`
	WalletBalance float64    `json:"wallet_balance"`
	TotalReports  int        `json:"total_reports"`
	TotalBookings int        `json:"total_bookings"`
	Streak        int        `json:"streak"`
	LastCheckIn   *time.Time `json:"last_check_in"`
	IsActive      bool       `json:"is_active"`
	CreatedAt     time.Time  `json:"created_at"`
}

func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:            u.ID,
		PhoneNumber:   u.PhoneNumber,
		Name:          u.Name,
		Ward:          u.Ward,
		Address:       u.Address,
		Points:        u.Points,
		WalletBalance: u.WalletBalance,
		TotalReports:  u.TotalReports,
		TotalBookings: u.TotalBookings,
		Streak:        u.Streak,
		LastCheckIn:   u.LastCheckIn,
		IsActive:      u.IsActive,
		CreatedAt:     u.CreatedAt,
	}
}
