CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone_number VARCHAR(15) UNIQUE NOT NULL,
    name VARCHAR(100),
    ward VARCHAR(50),
    address TEXT,
    points INTEGER DEFAULT 0,
    wallet_balance DECIMAL(10,2) DEFAULT 0,
    total_reports INTEGER DEFAULT 0,
    total_bookings INTEGER DEFAULT 0,
    streak INTEGER DEFAULT 0,
    last_check_in TIMESTAMPTZ,
    fcm_token TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_ward ON users(ward);
CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone_number);
