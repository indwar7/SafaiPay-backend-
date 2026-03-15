package report

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/indwar7/safaipay-backend/internal/notification"
	"github.com/indwar7/safaipay-backend/internal/user"
	"github.com/indwar7/safaipay-backend/pkg/storage"
	"gorm.io/gorm"
)

type Service interface {
	CreateReport(ctx context.Context, userID string, req *CreateReportRequest, imageURL string) (*Report, error)
	ListReports(ctx context.Context, filter *ReportFilter) ([]Report, int64, error)
	GetReport(ctx context.Context, id string) (*Report, error)
	UpdateStatus(ctx context.Context, id string, req *UpdateStatusRequest) (*Report, error)
}

type service struct {
	db              *gorm.DB
	userService     user.Service
	storageService  *storage.R2Service
	notifService    notification.Service
}

func NewService(db *gorm.DB, userSvc user.Service, storageSvc *storage.R2Service, notifSvc notification.Service) Service {
	return &service{
		db:             db,
		userService:    userSvc,
		storageService: storageSvc,
		notifService:   notifSvc,
	}
}

func (s *service) CreateReport(ctx context.Context, userID string, req *CreateReportRequest, imageURL string) (*Report, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	report := &Report{
		ID:           uuid.New(),
		UserID:       uid,
		IssueType:    req.IssueType,
		Description:  req.Description,
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
		Address:      req.Address,
		ImageURL:     imageURL,
		Status:       "pending",
		PointsEarned: 5,
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(report).Error; err != nil {
			return err
		}

		if err := s.userService.AddPoints(ctx, userID, 5, "Report submitted"); err != nil {
			return err
		}

		if err := s.userService.IncrementReports(ctx, userID); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	go func() {
		fcmToken, err := s.userService.GetFCMToken(context.Background(), userID)
		if err == nil && fcmToken != "" {
			n := notification.PointsEarned(5, "submitting a report")
			if err := s.notifService.Send(context.Background(), fcmToken, n); err != nil {
				slog.Error("failed to send points notification", "error", err)
			}
		}
	}()

	return report, nil
}

func (s *service) ListReports(ctx context.Context, filter *ReportFilter) ([]Report, int64, error) {
	var reports []Report
	var total int64

	query := s.db.WithContext(ctx).Model(&Report{})

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.IssueType != "" {
		query = query.Where("issue_type = ?", filter.IssueType)
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

	if err := query.Order("created_at DESC").Offset(offset).Limit(filter.Limit).Find(&reports).Error; err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

func (s *service) GetReport(ctx context.Context, id string) (*Report, error) {
	var report Report
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&report).Error; err != nil {
		return nil, err
	}
	return &report, nil
}

func (s *service) UpdateStatus(ctx context.Context, id string, req *UpdateStatusRequest) (*Report, error) {
	var report Report
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&report).Error; err != nil {
		return nil, err
	}

	updates := map[string]interface{}{
		"status": req.Status,
	}
	if req.ResolvedBy != "" {
		resolvedByUUID, err := uuid.Parse(req.ResolvedBy)
		if err != nil {
			return nil, fmt.Errorf("invalid resolved_by ID: %w", err)
		}
		updates["resolved_by"] = resolvedByUUID
	}

	if err := s.db.WithContext(ctx).Model(&report).Updates(updates).Error; err != nil {
		return nil, err
	}

	if req.Status == "resolved" {
		go func() {
			fcmToken, err := s.userService.GetFCMToken(context.Background(), report.UserID.String())
			if err == nil && fcmToken != "" {
				n := notification.ReportResolved(report.ID.String(), report.Address)
				if err := s.notifService.Send(context.Background(), fcmToken, n); err != nil {
					slog.Error("failed to send resolved notification", "error", err)
				}
			}
		}()
	}

	return &report, nil
}
