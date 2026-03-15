package badges

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/indwar7/safaipay-backend/internal/notification"
	"github.com/indwar7/safaipay-backend/internal/user"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Service interface {
	CheckAndAwardBadges(ctx context.Context, userID string) error
	GetUserBadges(ctx context.Context, userID string) ([]BadgeProgress, error)
	GetAllBadges(ctx context.Context) ([]Badge, error)
	SeedBadges(ctx context.Context) error
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

func (s *service) SeedBadges(ctx context.Context) error {
	badges := []Badge{
		{ID: uuid.New(), Name: "First Step", Description: "Submit your first report", Tier: "bronze", TriggerType: "reports_count", TriggerValue: 1, BonusPoints: 5},
		{ID: uuid.New(), Name: "Reporter", Description: "Submit 10 reports", Tier: "silver", TriggerType: "reports_count", TriggerValue: 10, BonusPoints: 20},
		{ID: uuid.New(), Name: "Clean Crusader", Description: "Submit 50 reports", Tier: "gold", TriggerType: "reports_count", TriggerValue: 50, BonusPoints: 100},
		{ID: uuid.New(), Name: "Check-in Champ", Description: "Maintain a 7-day streak", Tier: "bronze", TriggerType: "streak", TriggerValue: 7, BonusPoints: 10},
		{ID: uuid.New(), Name: "Streak Master", Description: "Maintain a 30-day streak", Tier: "gold", TriggerType: "streak", TriggerValue: 30, BonusPoints: 50},
		{ID: uuid.New(), Name: "Eco Warrior", Description: "Collect 100 kg of waste", Tier: "gold", TriggerType: "total_collected", TriggerValue: 100, BonusPoints: 100},
		{ID: uuid.New(), Name: "Point Millionaire", Description: "Earn 1000 points", Tier: "gold", TriggerType: "points", TriggerValue: 1000, BonusPoints: 50},
		{ID: uuid.New(), Name: "Community Hero", Description: "Reach top 10 on leaderboard", Tier: "gold", TriggerType: "leaderboard_rank", TriggerValue: 10, BonusPoints: 100},
		{ID: uuid.New(), Name: "Speed Reporter", Description: "Submit 5 reports in one day", Tier: "silver", TriggerType: "daily_reports", TriggerValue: 5, BonusPoints: 25},
		{ID: uuid.New(), Name: "First Pickup", Description: "Complete your first booking", Tier: "bronze", TriggerType: "bookings_count", TriggerValue: 1, BonusPoints: 5},
		{ID: uuid.New(), Name: "Waste Warrior", Description: "Complete 10 bookings", Tier: "silver", TriggerType: "bookings_count", TriggerValue: 10, BonusPoints: 25},
		{ID: uuid.New(), Name: "Generous", Description: "Make your first redemption", Tier: "bronze", TriggerType: "redemptions", TriggerValue: 1, BonusPoints: 5},
		{ID: uuid.New(), Name: "Big Spender", Description: "Redeem ₹500 worth of points", Tier: "silver", TriggerType: "total_redeemed", TriggerValue: 500, BonusPoints: 25},
		{ID: uuid.New(), Name: "Night Owl", Description: "Submit a report between 12am-4am", Tier: "silver", TriggerType: "night_report", TriggerValue: 1, BonusPoints: 15},
		{ID: uuid.New(), Name: "Top Earner", Description: "Earn 500 points in a week", Tier: "gold", TriggerType: "weekly_points", TriggerValue: 500, BonusPoints: 50},
	}

	for _, b := range badges {
		result := s.db.WithContext(ctx).Where("name = ?", b.Name).FirstOrCreate(&b)
		if result.Error != nil {
			return result.Error
		}
	}

	slog.Info("badges seeded", "count", len(badges))
	return nil
}

func (s *service) CheckAndAwardBadges(ctx context.Context, userID string) error {
	u, err := s.userService.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	var allBadges []Badge
	if err := s.db.WithContext(ctx).Find(&allBadges).Error; err != nil {
		return err
	}

	for _, badge := range allBadges {
		var currentValue int

		switch badge.TriggerType {
		case "reports_count":
			currentValue = u.TotalReports
		case "bookings_count":
			currentValue = u.TotalBookings
		case "streak":
			currentValue = u.Streak
		case "points":
			currentValue = u.Points
		default:
			continue
		}

		if currentValue >= badge.TriggerValue {
			s.awardBadge(ctx, userID, u.ID, badge)
		}
	}

	return nil
}

func (s *service) awardBadge(ctx context.Context, userIDStr string, userID uuid.UUID, badge Badge) {
	now := time.Now()
	ub := UserBadge{
		ID:         uuid.New(),
		UserID:     userID,
		BadgeID:    badge.ID,
		Progress:   badge.TriggerValue,
		UnlockedAt: &now,
	}

	result := s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "badge_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"progress"}),
	}).Create(&ub)

	if result.Error != nil {
		slog.Error("failed to award badge", "error", result.Error, "badge", badge.Name)
		return
	}

	if result.RowsAffected > 0 && badge.BonusPoints > 0 {
		if err := s.userService.AddPoints(ctx, userIDStr, badge.BonusPoints, "Badge unlocked: "+badge.Name); err != nil {
			slog.Error("failed to award badge bonus points", "error", err)
		}

		go func() {
			fcmToken, err := s.userService.GetFCMToken(context.Background(), userIDStr)
			if err == nil && fcmToken != "" {
				n := notification.BadgeUnlocked(badge.Name)
				if err := s.notifService.Send(context.Background(), fcmToken, n); err != nil {
					slog.Error("failed to send badge notification", "error", err)
				}
			}
		}()
	}
}

func (s *service) GetUserBadges(ctx context.Context, userID string) ([]BadgeProgress, error) {
	var allBadges []Badge
	if err := s.db.WithContext(ctx).Find(&allBadges).Error; err != nil {
		return nil, err
	}

	var userBadges []UserBadge
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Find(&userBadges).Error; err != nil {
		return nil, err
	}

	ubMap := make(map[uuid.UUID]UserBadge)
	for _, ub := range userBadges {
		ubMap[ub.BadgeID] = ub
	}

	result := make([]BadgeProgress, len(allBadges))
	for i, badge := range allBadges {
		bp := BadgeProgress{
			Badge: badge,
		}
		if ub, ok := ubMap[badge.ID]; ok {
			bp.Progress = ub.Progress
			bp.IsUnlocked = ub.UnlockedAt != nil
			bp.UnlockedAt = ub.UnlockedAt
		}
		result[i] = bp
	}

	return result, nil
}

func (s *service) GetAllBadges(ctx context.Context) ([]Badge, error) {
	var badges []Badge
	if err := s.db.WithContext(ctx).Find(&badges).Error; err != nil {
		return nil, err
	}
	return badges, nil
}
