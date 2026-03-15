package user

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service interface {
	GetOrCreateByPhone(ctx context.Context, phone string) (*User, error)
	GetProfile(ctx context.Context, userID string) (*UserResponse, error)
	UpdateProfile(ctx context.Context, userID string, req *UpdateProfileRequest) (*UserResponse, error)
	AddPoints(ctx context.Context, userID string, points int, description string) error
	DeductPoints(ctx context.Context, userID string, points int) error
	DailyCheckIn(ctx context.Context, userID string) (*UserResponse, error)
	GetFCMToken(ctx context.Context, userID string) (string, error)
	IncrementReports(ctx context.Context, userID string) error
	IncrementBookings(ctx context.Context, userID string) error
	CreditWallet(ctx context.Context, userID string, amount float64) error
	DebitWallet(ctx context.Context, userID string, amount float64) error
	GetUserByID(ctx context.Context, userID string) (*User, error)
}

type service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) Service {
	return &service{db: db}
}

func (s *service) GetOrCreateByPhone(ctx context.Context, phone string) (*User, error) {
	var u User
	result := s.db.WithContext(ctx).Where("phone_number = ?", phone).First(&u)
	if result.Error == nil {
		return &u, nil
	}

	if result.Error != gorm.ErrRecordNotFound {
		return nil, result.Error
	}

	u = User{
		ID:          uuid.New(),
		PhoneNumber: phone,
		IsActive:    true,
	}
	if err := s.db.WithContext(ctx).Create(&u).Error; err != nil {
		return nil, err
	}

	slog.Info("new user created", "phone", phone, "id", u.ID)
	return &u, nil
}

func (s *service) GetProfile(ctx context.Context, userID string) (*UserResponse, error) {
	var u User
	if err := s.db.WithContext(ctx).Where("id = ?", userID).First(&u).Error; err != nil {
		return nil, err
	}
	return u.ToResponse(), nil
}

func (s *service) UpdateProfile(ctx context.Context, userID string, req *UpdateProfileRequest) (*UserResponse, error) {
	var u User
	if err := s.db.WithContext(ctx).Where("id = ?", userID).First(&u).Error; err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Ward != nil {
		updates["ward"] = *req.Ward
	}
	if req.Address != nil {
		updates["address"] = *req.Address
	}
	if req.FCMToken != nil {
		updates["fcm_token"] = *req.FCMToken
	}

	if len(updates) > 0 {
		if err := s.db.WithContext(ctx).Model(&u).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	if err := s.db.WithContext(ctx).Where("id = ?", userID).First(&u).Error; err != nil {
		return nil, err
	}
	return u.ToResponse(), nil
}

func (s *service) AddPoints(ctx context.Context, userID string, points int, description string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&User{}).Where("id = ?", userID).
			Update("points", gorm.Expr("points + ?", points))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("user not found")
		}
		return nil
	})
}

func (s *service) DeductPoints(ctx context.Context, userID string, points int) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&User{}).Where("id = ? AND points >= ?", userID, points).
			Update("points", gorm.Expr("points - ?", points))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("insufficient points or user not found")
		}
		return nil
	})
}

func (s *service) DailyCheckIn(ctx context.Context, userID string) (*UserResponse, error) {
	var u User
	if err := s.db.WithContext(ctx).Where("id = ?", userID).First(&u).Error; err != nil {
		return nil, err
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if u.LastCheckIn != nil {
		lastDate := time.Date(u.LastCheckIn.Year(), u.LastCheckIn.Month(), u.LastCheckIn.Day(), 0, 0, 0, 0, u.LastCheckIn.Location())
		if lastDate.Equal(today) {
			return nil, fmt.Errorf("already checked in today")
		}

		yesterday := today.AddDate(0, 0, -1)
		if lastDate.Equal(yesterday) {
			u.Streak++
		} else {
			u.Streak = 1
		}
	} else {
		u.Streak = 1
	}

	u.LastCheckIn = &now
	u.Points += 2

	if err := s.db.WithContext(ctx).Model(&u).Updates(map[string]interface{}{
		"last_check_in": u.LastCheckIn,
		"streak":        u.Streak,
		"points":        u.Points,
	}).Error; err != nil {
		return nil, err
	}

	return u.ToResponse(), nil
}

func (s *service) GetFCMToken(ctx context.Context, userID string) (string, error) {
	var u User
	if err := s.db.WithContext(ctx).Select("fcm_token").Where("id = ?", userID).First(&u).Error; err != nil {
		return "", err
	}
	return u.FCMToken, nil
}

func (s *service) IncrementReports(ctx context.Context, userID string) error {
	return s.db.WithContext(ctx).Model(&User{}).Where("id = ?", userID).
		Update("total_reports", gorm.Expr("total_reports + 1")).Error
}

func (s *service) IncrementBookings(ctx context.Context, userID string) error {
	return s.db.WithContext(ctx).Model(&User{}).Where("id = ?", userID).
		Update("total_bookings", gorm.Expr("total_bookings + 1")).Error
}

func (s *service) CreditWallet(ctx context.Context, userID string, amount float64) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&User{}).Where("id = ?", userID).
			Update("wallet_balance", gorm.Expr("wallet_balance + ?", amount))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("user not found")
		}
		return nil
	})
}

func (s *service) DebitWallet(ctx context.Context, userID string, amount float64) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&User{}).Where("id = ? AND wallet_balance >= ?", userID, amount).
			Update("wallet_balance", gorm.Expr("wallet_balance - ?", amount))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("insufficient wallet balance or user not found")
		}
		return nil
	})
}

func (s *service) GetUserByID(ctx context.Context, userID string) (*User, error) {
	var u User
	if err := s.db.WithContext(ctx).Where("id = ?", userID).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}
