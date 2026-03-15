package payment

import (
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Type         string    `gorm:"size:20;not null;index" json:"type"`
	Points       int       `gorm:"default:0" json:"points"`
	Amount       float64   `gorm:"type:decimal(10,2);default:0" json:"amount"`
	Description  string    `json:"description"`
	Status       string    `gorm:"size:20;default:completed" json:"status"`
	RazorpayRef  string    `gorm:"size:100;column:razorpay_ref_id" json:"razorpay_ref_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type RedeemRequest struct {
	Points int `json:"points" validate:"required,min=1"`
}

type WithdrawRequest struct {
	Amount float64 `json:"amount" validate:"required,min=100,max=50000"`
}

type RazorpayVerifyRequest struct {
	PaymentID string `json:"payment_id" validate:"required"`
	OrderID   string `json:"order_id" validate:"required"`
	Signature string `json:"signature" validate:"required"`
}

type TransactionFilter struct {
	Type  string `form:"type"`
	Page  int    `form:"page,default=1"`
	Limit int    `form:"limit,default=20"`
}

type WalletResponse struct {
	Points        int     `json:"points"`
	WalletBalance float64 `json:"wallet_balance"`
}
