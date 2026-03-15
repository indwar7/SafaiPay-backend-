package booking

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/indwar7/safaipay-backend/internal/notification"
	"github.com/indwar7/safaipay-backend/internal/user"
	"gorm.io/gorm"
)

type Service interface {
	CreateBooking(ctx context.Context, userID string, req *CreateBookingRequest) (*Booking, error)
	GetUserBookings(ctx context.Context, userID string, filter *BookingFilter) ([]Booking, int64, error)
	GetBooking(ctx context.Context, id string) (*Booking, error)
	UpdateStatus(ctx context.Context, id string, req *UpdateBookingStatusRequest) (*Booking, error)
	GetCollectorBookings(ctx context.Context, collectorID string, filter *BookingFilter) ([]Booking, int64, error)
	CompleteBooking(ctx context.Context, bookingID, collectorID string, weight float64, imageURL string) (*Booking, error)
	AssignCollector(ctx context.Context, bookingID string, collectorID uuid.UUID) error
}

type service struct {
	db           *gorm.DB
	userService  user.Service
	notifService notification.Service
}

func NewService(db *gorm.DB, userSvc user.Service, notifSvc notification.Service) Service {
	return &service{
		db:           db,
		userService:  userSvc,
		notifService: notifSvc,
	}
}

func (s *service) CreateBooking(ctx context.Context, userID string, req *CreateBookingRequest) (*Booking, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	bookingDate, err := time.Parse(time.RFC3339, req.BookingDate)
	if err != nil {
		return nil, fmt.Errorf("invalid booking date format, use ISO 8601: %w", err)
	}

	booking := &Booking{
		ID:          uuid.New(),
		UserID:      uid,
		WasteType:   req.WasteType,
		BookingDate: bookingDate,
		TimeSlot:    req.TimeSlot,
		Address:     req.Address,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Status:      "pending",
	}

	if err := s.db.WithContext(ctx).Create(booking).Error; err != nil {
		return nil, err
	}

	if err := s.userService.IncrementBookings(ctx, userID); err != nil {
		slog.Error("failed to increment bookings count", "error", err)
	}

	return booking, nil
}

func (s *service) GetUserBookings(ctx context.Context, userID string, filter *BookingFilter) ([]Booking, int64, error) {
	var bookings []Booking
	var total int64

	query := s.db.WithContext(ctx).Model(&Booking{}).Where("user_id = ?", userID)

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
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

	if err := query.Order("created_at DESC").Offset(offset).Limit(filter.Limit).Find(&bookings).Error; err != nil {
		return nil, 0, err
	}

	return bookings, total, nil
}

func (s *service) GetBooking(ctx context.Context, id string) (*Booking, error) {
	var booking Booking
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&booking).Error; err != nil {
		return nil, err
	}
	return &booking, nil
}

func (s *service) UpdateStatus(ctx context.Context, id string, req *UpdateBookingStatusRequest) (*Booking, error) {
	var booking Booking
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&booking).Error; err != nil {
		return nil, err
	}

	booking.Status = req.Status
	if err := s.db.WithContext(ctx).Save(&booking).Error; err != nil {
		return nil, err
	}

	return &booking, nil
}

func (s *service) GetCollectorBookings(ctx context.Context, collectorID string, filter *BookingFilter) ([]Booking, int64, error) {
	var bookings []Booking
	var total int64

	query := s.db.WithContext(ctx).Model(&Booking{}).Where("collector_id = ?", collectorID)

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
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

	if err := query.Order("created_at DESC").Offset(offset).Limit(filter.Limit).Find(&bookings).Error; err != nil {
		return nil, 0, err
	}

	return bookings, total, nil
}

func (s *service) CompleteBooking(ctx context.Context, bookingID, collectorID string, weight float64, imageURL string) (*Booking, error) {
	var booking Booking

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ? AND collector_id = ?", bookingID, collectorID).First(&booking).Error; err != nil {
			return fmt.Errorf("booking not found or not assigned to this collector")
		}

		points := int(weight * 10)
		booking.Status = "completed"
		booking.Weight = &weight
		booking.PointsEarned = points
		booking.ImageURL = imageURL

		if err := tx.Save(&booking).Error; err != nil {
			return err
		}

		if err := s.userService.AddPoints(ctx, booking.UserID.String(), points, fmt.Sprintf("Pickup completed: %.2f kg", weight)); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	go func() {
		fcmToken, err := s.userService.GetFCMToken(context.Background(), booking.UserID.String())
		if err == nil && fcmToken != "" {
			points := int(weight * 10)
			n := notification.PointsEarned(points, fmt.Sprintf("pickup of %.2f kg", weight))
			if err := s.notifService.Send(context.Background(), fcmToken, n); err != nil {
				slog.Error("failed to send completion notification", "error", err)
			}
		}
	}()

	return &booking, nil
}

func (s *service) AssignCollector(ctx context.Context, bookingID string, collectorID uuid.UUID) error {
	return s.db.WithContext(ctx).Model(&Booking{}).Where("id = ?", bookingID).
		Updates(map[string]interface{}{
			"collector_id": collectorID,
			"status":       "assigned",
		}).Error
}
