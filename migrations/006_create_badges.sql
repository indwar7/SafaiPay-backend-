CREATE TABLE IF NOT EXISTS badges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    icon_url TEXT,
    tier VARCHAR(20) NOT NULL,
    trigger_type VARCHAR(50) NOT NULL,
    trigger_value INTEGER NOT NULL,
    bonus_points INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_badges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    badge_id UUID NOT NULL REFERENCES badges(id),
    progress INTEGER DEFAULT 0,
    unlocked_at TIMESTAMPTZ,
    UNIQUE(user_id, badge_id)
);

CREATE INDEX IF NOT EXISTS idx_user_badges_user_id ON user_badges(user_id);

-- Seed default badges
INSERT INTO badges (name, description, icon_url, tier, trigger_type, trigger_value, bonus_points) VALUES
    ('First Step', 'Submit your first report', '', 'bronze', 'reports_count', 1, 5),
    ('Reporter', 'Submit 10 reports', '', 'silver', 'reports_count', 10, 20),
    ('Clean Crusader', 'Submit 50 reports', '', 'gold', 'reports_count', 50, 100),
    ('Check-in Champ', 'Maintain a 7-day streak', '', 'bronze', 'streak', 7, 10),
    ('Streak Master', 'Maintain a 30-day streak', '', 'gold', 'streak', 30, 50),
    ('Eco Warrior', 'Collect 100 kg of waste', '', 'gold', 'total_collected', 100, 100),
    ('Point Millionaire', 'Earn 1000 points', '', 'gold', 'points', 1000, 50),
    ('Community Hero', 'Reach top 10 on leaderboard', '', 'gold', 'leaderboard_rank', 10, 100),
    ('Speed Reporter', 'Submit 5 reports in one day', '', 'silver', 'daily_reports', 5, 25),
    ('First Pickup', 'Complete your first booking', '', 'bronze', 'bookings_count', 1, 5),
    ('Waste Warrior', 'Complete 10 bookings', '', 'silver', 'bookings_count', 10, 25),
    ('Generous', 'Make your first redemption', '', 'bronze', 'redemptions', 1, 5),
    ('Big Spender', 'Redeem ₹500 worth of points', '', 'silver', 'total_redeemed', 500, 25),
    ('Night Owl', 'Submit a report between 12am-4am', '', 'silver', 'night_report', 1, 15),
    ('Top Earner', 'Earn 500 points in a week', '', 'gold', 'weekly_points', 500, 50)
ON CONFLICT DO NOTHING;
