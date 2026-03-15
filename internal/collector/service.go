package collector

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service interface {
	GetOrCreateByPhone(ctx context.Context, phone string) (*Collector, error)
	GetProfile(ctx context.Context, collectorID string) (*CollectorResponse, error)
	UpdateLocation(ctx context.Context, collectorID string, lat, lng float64) error
	FindNearestAvailable(ctx context.Context, ward string, lat, lng float64) (*Collector, error)
	UpdateStatus(ctx context.Context, collectorID, status string) error
	AddCollected(ctx context.Context, collectorID string, weight float64) error
	GetFCMToken(ctx context.Context, collectorID string) (string, error)
}

type service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) Service {
	return &service{db: db}
}

func (s *service) GetOrCreateByPhone(ctx context.Context, phone string) (*Collector, error) {
	var c Collector
	result := s.db.WithContext(ctx).Where("phone_number = ?", phone).First(&c)
	if result.Error == nil {
		return &c, nil
	}

	if result.Error != gorm.ErrRecordNotFound {
		return nil, result.Error
	}

	c = Collector{
		ID:          uuid.New(),
		PhoneNumber: phone,
		Status:      "available",
		Rating:      5.0,
	}
	if err := s.db.WithContext(ctx).Create(&c).Error; err != nil {
		return nil, err
	}

	slog.Info("new collector created", "phone", phone, "id", c.ID)
	return &c, nil
}

func (s *service) GetProfile(ctx context.Context, collectorID string) (*CollectorResponse, error) {
	var c Collector
	if err := s.db.WithContext(ctx).Where("id = ?", collectorID).First(&c).Error; err != nil {
		return nil, err
	}
	return c.ToResponse(), nil
}

func (s *service) UpdateLocation(ctx context.Context, collectorID string, lat, lng float64) error {
	result := s.db.WithContext(ctx).Model(&Collector{}).Where("id = ?", collectorID).
		Updates(map[string]interface{}{
			"current_lat": lat,
			"current_lng": lng,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("collector not found")
	}
	return nil
}

func (s *service) FindNearestAvailable(ctx context.Context, ward string, lat, lng float64) (*Collector, error) {
	var c Collector

	query := s.db.WithContext(ctx).
		Where("status = ? AND is_verified = ?", "available", true)

	if ward != "" {
		query = query.Where("ward = ?", ward)
	}

	// Order by distance using Euclidean approximation
	// For precise distance, PostGIS would be better, but this works for nearby matches
	query = query.Order(fmt.Sprintf(
		"((current_lat - %f) * (current_lat - %f) + (current_lng - %f) * (current_lng - %f)) ASC",
		lat, lat, lng, lng,
	))

	if err := query.First(&c).Error; err != nil {
		return nil, err
	}

	return &c, nil
}

func (s *service) UpdateStatus(ctx context.Context, collectorID, status string) error {
	return s.db.WithContext(ctx).Model(&Collector{}).Where("id = ?", collectorID).
		Update("status", status).Error
}

func (s *service) AddCollected(ctx context.Context, collectorID string, weight float64) error {
	return s.db.WithContext(ctx).Model(&Collector{}).Where("id = ?", collectorID).
		Update("total_collected", gorm.Expr("total_collected + ?", weight)).Error
}

func (s *service) GetFCMToken(ctx context.Context, collectorID string) (string, error) {
	var c Collector
	if err := s.db.WithContext(ctx).Select("fcm_token").Where("id = ?", collectorID).First(&c).Error; err != nil {
		return "", err
	}
	return c.FCMToken, nil
}
