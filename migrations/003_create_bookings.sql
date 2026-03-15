CREATE TABLE IF NOT EXISTS bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    collector_id UUID,
    waste_type VARCHAR(50) NOT NULL,
    booking_date TIMESTAMPTZ NOT NULL,
    time_slot VARCHAR(30) NOT NULL,
    address TEXT NOT NULL,
    latitude DECIMAL(10,8),
    longitude DECIMAL(11,8),
    status VARCHAR(20) DEFAULT 'pending',
    weight DECIMAL(6,2),
    points_earned INTEGER DEFAULT 0,
    image_url TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_bookings_user_id ON bookings(user_id);
CREATE INDEX IF NOT EXISTS idx_bookings_collector_id ON bookings(collector_id);
CREATE INDEX IF NOT EXISTS idx_bookings_status ON bookings(status);
