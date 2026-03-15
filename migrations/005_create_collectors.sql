CREATE TABLE IF NOT EXISTS collectors (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    phone_number VARCHAR(15) UNIQUE NOT NULL,
    name VARCHAR(100),
    ward VARCHAR(50),
    current_lat DECIMAL(10,8),
    current_lng DECIMAL(11,8),
    status VARCHAR(20) DEFAULT 'available',
    rating DECIMAL(3,2) DEFAULT 5.0,
    total_collected DECIMAL(10,2) DEFAULT 0,
    fcm_token TEXT,
    bank_account_number VARCHAR(20),
    bank_ifsc VARCHAR(15),
    bank_name VARCHAR(100),
    is_verified BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_collectors_ward ON collectors(ward);
CREATE INDEX IF NOT EXISTS idx_collectors_status ON collectors(status);

-- Add foreign key to reports.resolved_by now that collectors table exists
ALTER TABLE reports ADD CONSTRAINT fk_reports_resolved_by
    FOREIGN KEY (resolved_by) REFERENCES collectors(id);

-- Add foreign key to bookings.collector_id
ALTER TABLE bookings ADD CONSTRAINT fk_bookings_collector_id
    FOREIGN KEY (collector_id) REFERENCES collectors(id);
