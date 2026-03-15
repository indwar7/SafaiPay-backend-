package payment

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/indwar7/safaipay-backend/config"
	"github.com/indwar7/safaipay-backend/internal/notification"
	"github.com/indwar7/safaipay-backend/internal/user"
	"gorm.io/gorm"
)

type Service interface {
	GetWallet(ctx context.Context, userID string) (*WalletResponse, error)
	RedeemPoints(ctx context.Context, userID string, req *RedeemRequest) (*Transaction, error)
	Withdraw(ctx context.Context, userID string, req *WithdrawRequest) (*Transaction, error)
	VerifyRazorpay(ctx context.Context, req *RazorpayVerifyRequest) error
	GetTransactions(ctx context.Context, userID string, filter *TransactionFilter) ([]Transaction, int64, error)
	LogTransaction(ctx context.Context, userID string, txType string, points int, amount float64, description, status, razorpayRef string) (*Transaction, error)
}

type service struct {
	db             *gorm.DB
	userService    user.Service
	notifService   notification.Service
	razorpayKeyID  string
	razorpaySecret string
	httpClient     *http.Client
}

func NewService(db *gorm.DB, userSvc user.Service, notifSvc notification.Service, cfg *config.RazorpayConfig) Service {
	return &service{
		db:             db,
		userService:    userSvc,
		notifService:   notifSvc,
		razorpayKeyID:  cfg.KeyID,
		razorpaySecret: cfg.KeySecret,
		httpClient:     &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *service) GetWallet(ctx context.Context, userID string) (*WalletResponse, error) {
	u, err := s.userService.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &WalletResponse{
		Points:        u.Points,
		WalletBalance: u.WalletBalance,
	}, nil
}

func (s *service) RedeemPoints(ctx context.Context, userID string, req *RedeemRequest) (*Transaction, error) {
	if req.Points < 1 {
		return nil, fmt.Errorf("minimum 1 point required for redemption")
	}

	var tx *Transaction

	err := s.db.WithContext(ctx).Transaction(func(dbTx *gorm.DB) error {
		if err := s.userService.DeductPoints(ctx, userID, req.Points); err != nil {
			return fmt.Errorf("deduct points: %w", err)
		}

		amount := float64(req.Points)
		if err := s.userService.CreditWallet(ctx, userID, amount); err != nil {
			return fmt.Errorf("credit wallet: %w", err)
		}

		uid, _ := uuid.Parse(userID)
		tx = &Transaction{
			ID:          uuid.New(),
			UserID:      uid,
			Type:        "redeemed",
			Points:      req.Points,
			Amount:      amount,
			Description: fmt.Sprintf("Redeemed %d points to wallet", req.Points),
			Status:      "completed",
		}
		return dbTx.Create(tx).Error
	})
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (s *service) Withdraw(ctx context.Context, userID string, req *WithdrawRequest) (*Transaction, error) {
	if req.Amount < 100 {
		return nil, fmt.Errorf("minimum withdrawal amount is ₹100")
	}
	if req.Amount > 50000 {
		return nil, fmt.Errorf("maximum withdrawal amount is ₹50,000 per day")
	}

	if err := s.userService.DebitWallet(ctx, userID, req.Amount); err != nil {
		return nil, fmt.Errorf("debit wallet: %w", err)
	}

	// Create payout via RazorpayX Payouts API (raw HTTP)
	payoutData := map[string]interface{}{
		"account_number":       "YOUR_RAZORPAY_ACCOUNT",
		"amount":               int(req.Amount * 100), // amount in paise
		"currency":             "INR",
		"mode":                 "IMPS",
		"purpose":              "payout",
		"fund_account_id":      "",
		"queue_if_low_balance": true,
	}

	payloadBytes, _ := json.Marshal(payoutData)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.razorpay.com/v1/payouts", bytes.NewReader(payloadBytes))
	if err != nil {
		s.refundWalletOnFailure(ctx, userID, req.Amount)
		return nil, fmt.Errorf("withdrawal failed, please try again")
	}
	httpReq.SetBasicAuth(s.razorpayKeyID, s.razorpaySecret)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		slog.Error("razorpay payout request failed", "error", err)
		s.refundWalletOnFailure(ctx, userID, req.Amount)
		return nil, fmt.Errorf("withdrawal failed, please try again")
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		slog.Error("razorpay payout failed", "status", resp.StatusCode, "body", string(respBody))
		s.refundWalletOnFailure(ctx, userID, req.Amount)
		return nil, fmt.Errorf("withdrawal failed, please try again")
	}

	var payoutResp map[string]interface{}
	json.Unmarshal(respBody, &payoutResp)
	payoutID, _ := payoutResp["id"].(string)

	uid, _ := uuid.Parse(userID)
	tx := &Transaction{
		ID:          uuid.New(),
		UserID:      uid,
		Type:        "withdrawn",
		Amount:      req.Amount,
		Description: fmt.Sprintf("Withdrawal of ₹%.2f to bank", req.Amount),
		Status:      "pending",
		RazorpayRef: payoutID,
	}
	if err := s.db.WithContext(ctx).Create(tx).Error; err != nil {
		slog.Error("failed to log withdrawal transaction", "error", err)
	}

	go func() {
		fcmToken, err := s.userService.GetFCMToken(context.Background(), userID)
		if err == nil && fcmToken != "" {
			n := notification.WithdrawalSuccess(req.Amount)
			if sendErr := s.notifService.Send(context.Background(), fcmToken, n); sendErr != nil {
				slog.Error("failed to send withdrawal notification", "error", sendErr)
			}
		}
	}()

	return tx, nil
}

func (s *service) refundWalletOnFailure(ctx context.Context, userID string, amount float64) {
	if creditErr := s.userService.CreditWallet(ctx, userID, amount); creditErr != nil {
		slog.Error("CRITICAL: failed to refund wallet after payout failure",
			"error", creditErr, "userID", userID, "amount", amount)
	}
}

func (s *service) VerifyRazorpay(ctx context.Context, req *RazorpayVerifyRequest) error {
	message := req.OrderID + "|" + req.PaymentID
	mac := hmac.New(sha256.New, []byte(s.razorpaySecret))
	mac.Write([]byte(message))
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(req.Signature)) {
		return fmt.Errorf("invalid payment signature")
	}

	return nil
}

func (s *service) GetTransactions(ctx context.Context, userID string, filter *TransactionFilter) ([]Transaction, int64, error) {
	var transactions []Transaction
	var total int64

	query := s.db.WithContext(ctx).Model(&Transaction{}).Where("user_id = ?", userID)

	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	offset := (filter.Page - 1) * filter.Limit

	if err := query.Order("created_at DESC").Offset(offset).Limit(filter.Limit).Find(&transactions).Error; err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

func (s *service) LogTransaction(ctx context.Context, userID string, txType string, points int, amount float64, description, status, razorpayRef string) (*Transaction, error) {
	uid, _ := uuid.Parse(userID)
	tx := &Transaction{
		ID:          uuid.New(),
		UserID:      uid,
		Type:        txType,
		Points:      points,
		Amount:      amount,
		Description: description,
		Status:      status,
		RazorpayRef: razorpayRef,
	}
	if err := s.db.WithContext(ctx).Create(tx).Error; err != nil {
		return nil, err
	}
	return tx, nil
}
