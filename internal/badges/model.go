package badges

import (
	"time"

	"github.com/google/uuid"
)

type Badge struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Name         string    `gorm:"size:100;not null" json:"name"`
	Description  string    `json:"description"`
	IconURL      string    `json:"icon_url"`
	Tier         string    `gorm:"size:20;not null" json:"tier"`
	TriggerType  string    `gorm:"size:50;not null" json:"trigger_type"`
	TriggerValue int       `gorm:"not null" json:"trigger_value"`
	BonusPoints  int       `gorm:"default:0" json:"bonus_points"`
	CreatedAt    time.Time `json:"created_at"`
}

type UserBadge struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID     uuid.UUID  `gorm:"type:uuid;not null;index;uniqueIndex:idx_user_badge" json:"user_id"`
	BadgeID    uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_user_badge" json:"badge_id"`
	Progress   int        `gorm:"default:0" json:"progress"`
	UnlockedAt *time.Time `json:"unlocked_at"`
	Badge      Badge      `gorm:"foreignKey:BadgeID" json:"badge"`
}

type BadgeProgress struct {
	Badge      Badge  `json:"badge"`
	Progress   int    `json:"progress"`
	IsUnlocked bool   `json:"is_unlocked"`
	UnlockedAt *time.Time `json:"unlocked_at,omitempty"`
}
