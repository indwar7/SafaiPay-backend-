# SafaiPay Backend

Production-ready Go backend for SafaiPay — a civic-tech app where citizens report cleanliness issues, book garbage pickups, and earn reward points redeemable as real money.

## Tech Stack

- **Language:** Go 1.22+
- **Framework:** Gin
- **Database:** PostgreSQL 16 (via GORM)
- **Cache:** Redis 7
- **Auth:** Phone OTP (MSG91) + JWT
- **Storage:** Cloudflare R2 (S3-compatible)
- **Payments:** Razorpay (RazorpayX for payouts)
- **Notifications:** FCM HTTP v1 API

## Quick Start

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- PostgreSQL 16
- Redis 7

### 1. Clone & Configure

```bash
git clone https://github.com/indwar7/safaipay-backend.git
cd safaipay-backend
cp .env.example .env
# Edit .env with your actual credentials
```

### 2. Run with Docker Compose

```bash
docker-compose up --build
```

This starts the app on `:8080` along with PostgreSQL and Redis.

### 3. Run Locally (without Docker)

```bash
# Start PostgreSQL and Redis separately, then:
go mod download
go run cmd/main.go
```

### 4. Health Check

```bash
curl http://localhost:8080/health
```

## Project Structure

```
safaipay-backend/
├── cmd/main.go                  # Entry point, wires all dependencies
├── config/config.go             # Environment configuration
├── internal/
│   ├── auth/                    # OTP send/verify, JWT generation
│   ├── user/                    # Profile, check-in, points, wallet
│   ├── report/                  # Issue reporting with image upload
│   ├── booking/                 # Garbage pickup bookings
│   ├── payment/                 # Points redemption, bank withdrawals
│   ├── leaderboard/             # Redis-based global & ward leaderboards
│   ├── badges/                  # 15 achievement badges with auto-award
│   ├── collector/               # Waste collector management
│   └── notification/            # FCM push notifications
├── pkg/
│   ├── database/                # PostgreSQL & Redis connections
│   ├── middleware/               # JWT auth, rate limiting, CORS
│   ├── sms/                     # MSG91 OTP integration
│   ├── storage/                 # Cloudflare R2 file uploads
│   └── response/                # Standardized JSON responses
├── migrations/                  # SQL migration files (001-006)
├── Dockerfile                   # Multi-stage production build
└── docker-compose.yml           # App + PostgreSQL + Redis
```

## API Endpoints

All responses follow the format:
```json
{
  "success": true,
  "message": "description",
  "data": {},
  "error": ""
}
```

### Auth (Public)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/send-otp` | Send OTP to phone (rate limited: 5/15min) |
| POST | `/api/v1/auth/verify-otp` | Verify OTP, returns JWT |

### User (Protected)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/user/profile` | Get user profile |
| PATCH | `/api/v1/user/profile` | Update name, ward, address, FCM token |
| POST | `/api/v1/user/checkin` | Daily check-in (+2 points, streak tracking) |

### Reports (Protected)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/reports` | Create report with image (multipart, +5 points) |
| GET | `/api/v1/reports` | List reports (filter: `status`, `issue_type`, `page`, `limit`) |
| GET | `/api/v1/reports/:id` | Get single report |
| PATCH | `/api/v1/reports/:id/status` | Update report status |

### Bookings (Protected)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/bookings` | Create pickup booking |
| GET | `/api/v1/bookings` | List user bookings |
| GET | `/api/v1/bookings/:id` | Get single booking |
| PATCH | `/api/v1/bookings/:id/status` | Update booking status |

### Wallet & Payments (Protected)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/wallet` | Get points + wallet balance |
| POST | `/api/v1/wallet/redeem` | Redeem points to wallet (1 point = ₹1) |
| POST | `/api/v1/wallet/withdraw` | Withdraw to bank (min ₹100, max ₹50,000/day) |
| POST | `/api/v1/payment/verify` | Verify Razorpay payment signature |
| GET | `/api/v1/wallet/transactions` | Transaction history (filter: `type`, `page`, `limit`) |

### Leaderboard (Protected)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/leaderboard` | Global or ward leaderboard (`?ward=xxx&limit=100`) |
| GET | `/api/v1/leaderboard/me` | Current user's rank |

### Badges (Protected)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/badges` | All available badges |
| GET | `/api/v1/badges/user` | User's badges with progress |

### Collector (Separate Auth)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/collector/send-otp` | Collector OTP login |
| POST | `/api/v1/collector/verify-otp` | Verify collector OTP |
| GET | `/api/v1/collector/profile` | Collector profile |
| GET | `/api/v1/collector/bookings` | Assigned bookings |
| PATCH | `/api/v1/collector/bookings/:id/complete` | Complete booking (weight + photo) |
| PATCH | `/api/v1/collector/location` | Update GPS location |

## Points System

| Action | Points |
|--------|--------|
| Daily check-in | +2 |
| Report an issue | +5 |
| Completed pickup | +10 per kg |
| Badge bonus | varies (5-100) |

- 1 point = ₹1
- Minimum withdrawal: ₹100
- Maximum withdrawal: ₹50,000/day

## Authentication

All protected endpoints require:
```
Authorization: Bearer <jwt_token>
```

Tokens are obtained via the OTP verify endpoint. User and collector tokens use separate roles.

## Environment Variables

See [.env.example](.env.example) for the full list:

- `SERVER_PORT` — HTTP port (default: 8080)
- `DB_*` — PostgreSQL connection
- `REDIS_*` — Redis connection
- `JWT_SECRET` — JWT signing key
- `MSG91_*` — SMS OTP provider
- `R2_*` — Cloudflare R2 storage
- `RAZORPAY_*` — Payment gateway
- `FCM_*` — Firebase Cloud Messaging

## Database Migrations

SQL migrations are in `migrations/` and are auto-applied via GORM AutoMigrate on startup. For manual execution:

```bash
psql -U safaipay -d safaipay -f migrations/001_create_users.sql
psql -U safaipay -d safaipay -f migrations/002_create_reports.sql
psql -U safaipay -d safaipay -f migrations/003_create_bookings.sql
psql -U safaipay -d safaipay -f migrations/004_create_transactions.sql
psql -U safaipay -d safaipay -f migrations/005_create_collectors.sql
psql -U safaipay -d safaipay -f migrations/006_create_badges.sql
```

## Architecture

- **Dependency Injection** — all services are behind interfaces, wired in `cmd/main.go`
- **DB Transactions** — all financial operations (points, wallet) are atomic
- **Rate Limiting** — Redis-based, per-phone on OTP endpoints
- **Leaderboard** — Redis sorted sets for O(log N) rank operations
- **Notifications** — FCM HTTP v1 API with OAuth2 token caching (no Firebase SDK)
- **Payouts** — RazorpayX API via raw HTTP with automatic wallet refund on failure

## License

Private — All rights reserved.
