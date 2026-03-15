# SafaiPay — Backend + Flutter Integration Guide

## What Has Been Built (Backend - COMPLETE)

A production-ready Go backend with **41 files**, fully compiled and pushed to GitHub.

### Backend Architecture

```
Flutter App  ──HTTP/JSON──▶  Go Backend (Gin)  ──▶  PostgreSQL
                                   │                    Redis
                                   │                    Cloudflare R2
                                   ├──▶ MSG91 (OTP SMS)
                                   ├──▶ Razorpay (Payouts)
                                   └──▶ FCM (Push Notifications)
```

### What Each Module Does

| Module | Status | What It Does |
|--------|--------|-------------|
| **Auth** | Done | Phone OTP login via MSG91, JWT token generation, separate user/collector auth |
| **User** | Done | Profile CRUD, daily check-in with streak tracking, atomic point/wallet operations |
| **Reports** | Done | Issue reporting with image upload to Cloudflare R2, +5 points auto-award |
| **Bookings** | Done | Garbage pickup scheduling, collector assignment, weight-based point calculation |
| **Payment** | Done | Points → wallet redemption (1pt = ₹1), bank withdrawal via RazorpayX, HMAC signature verification |
| **Leaderboard** | Done | Redis sorted sets, global + ward-level rankings, O(log N) operations |
| **Badges** | Done | 15 achievement badges, auto-check after user actions, bonus point awards |
| **Collector** | Done | Collector auth, nearest-collector assignment, GPS location updates, booking completion |
| **Notifications** | Done | FCM HTTP v1 push notifications, 9 notification templates |
| **Middleware** | Done | JWT auth, Redis rate limiting (5 OTP/15min/phone), CORS |
| **Database** | Done | PostgreSQL with GORM, Redis, connection pooling, 6 migration files |
| **Storage** | Done | Cloudflare R2 image uploads for reports, bookings, profiles |

---

## What the Frontend Needs to Do

### Flutter Packages Required

```yaml
# pubspec.yaml
dependencies:
  dio: ^5.4.0              # HTTP client (recommended over http package)
  flutter_secure_storage: ^9.0.0  # Store JWT token securely
  image_picker: ^1.0.7     # Camera/gallery for report images
  geolocator: ^11.0.0      # GPS location for reports & bookings
  firebase_messaging: ^14.7.0    # FCM push notifications
  firebase_core: ^2.25.0
  google_maps_flutter: ^2.5.0   # Map for report location
  razorpay_flutter: ^1.3.7      # Payment (if needed for top-up)
  cached_network_image: ^3.3.1  # Load report/badge images
  intl: ^0.19.0            # Date formatting
  provider: ^6.1.1         # State management (or riverpod/bloc)
```

---

### Screen-by-Screen Integration

#### 1. Login / OTP Screen

**Flow:** Enter phone → Send OTP → Enter OTP → Get JWT → Save token → Navigate to Home

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  Enter Phone     │────▶│  Enter OTP      │────▶│  Home Screen    │
│  (+91XXXXXXXXXX) │     │  (6 digits)     │     │                 │
└─────────────────┘     └─────────────────┘     └─────────────────┘
```

**API Calls:**

```dart
// Step 1: Send OTP
POST /api/v1/auth/send-otp
Body: { "phone_number": "+919876543210" }
Response: { "success": true, "message": "OTP sent successfully" }

// Step 2: Verify OTP
POST /api/v1/auth/verify-otp
Body: { "phone_number": "+919876543210", "otp": "123456" }
Response: {
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",    // ← SAVE THIS
    "user": {
      "id": "uuid",
      "phone_number": "+919876543210",
      "name": "",
      "points": 0,
      "wallet_balance": 0.00,
      "streak": 0
    }
  }
}
```

**Frontend Tasks:**
- [ ] Store JWT in `flutter_secure_storage`
- [ ] Phone number must be E.164 format: `+91XXXXXXXXXX`
- [ ] Add JWT to ALL subsequent requests: `Authorization: Bearer <token>`
- [ ] Handle rate limit (429) — show "Try again in X minutes"
- [ ] After login, call `PATCH /user/profile` to send FCM token

---

#### 2. Home Screen

**Shows:** User name, points, streak, check-in button, quick actions

**API Calls:**

```dart
// Get profile (on load)
GET /api/v1/user/profile
Headers: { "Authorization": "Bearer <token>" }
Response: {
  "data": {
    "name": "Abhay",
    "points": 150,
    "wallet_balance": 50.00,
    "streak": 7,
    "total_reports": 12,
    "total_bookings": 3,
    "last_check_in": "2026-03-15T10:30:00Z"
  }
}

// Daily check-in (button tap)
POST /api/v1/user/checkin
Response: { "message": "check-in successful! +2 points", "data": { ... } }
// Error if already checked in: { "success": false, "message": "already checked in today" }
```

**Frontend Tasks:**
- [ ] Show check-in button (disabled if `last_check_in` is today)
- [ ] Animate streak counter and points
- [ ] Show "+2" toast on successful check-in
- [ ] Pull-to-refresh to reload profile

---

#### 3. Report Issue Screen

**Flow:** Choose issue type → Take photo → Pick location → Add description → Submit

**API Call:**

```dart
// Create report (multipart form data — NOT JSON)
POST /api/v1/reports
Content-Type: multipart/form-data

Fields:
  issue_type: "Overflowing Bin"          // required
  description: "Bin has been full for 3 days"
  latitude: "19.0760"                    // from GPS
  longitude: "72.8777"
  address: "MG Road, Mumbai"
  image: <file>                          // from camera/gallery

Response: {
  "success": true,
  "message": "report created successfully, +5 points earned!",
  "data": {
    "id": "uuid",
    "issue_type": "Overflowing Bin",
    "image_url": "https://pub-xxx.r2.dev/reports/userId/image.jpg",
    "status": "pending",
    "points_earned": 5
  }
}
```

**Issue Types (use these exact strings):**
- `Overflowing Bin`
- `Dirty Street`
- `Illegal Dumping`
- `Broken Drain`
- `Dead Animal`
- `Construction Waste`
- `Other`

**Frontend Tasks:**
- [ ] Use `image_picker` for camera/gallery
- [ ] Use `geolocator` to get current lat/lng
- [ ] Use `Dio` with `FormData` for multipart upload
- [ ] Show "+5 points" animation on success
- [ ] Compress image before upload (keep under 5MB)

---

#### 4. My Reports Screen

**API Calls:**

```dart
// List reports with pagination
GET /api/v1/reports?status=pending&page=1&limit=20

// Filter options: status=pending|assigned|resolved, issue_type=xxx

Response: {
  "data": {
    "reports": [ { "id": "...", "issue_type": "...", "status": "pending", "image_url": "...", "created_at": "..." } ],
    "total": 45,
    "page": 1,
    "limit": 20
  }
}

// Get single report
GET /api/v1/reports/:id
```

**Frontend Tasks:**
- [ ] Infinite scroll / pagination (increment `page`)
- [ ] Filter chips: All, Pending, Assigned, Resolved
- [ ] Show report image with `cached_network_image`
- [ ] Color-coded status badges (yellow/blue/green)

---

#### 5. Book Pickup Screen

**Flow:** Select waste type → Pick date/time slot → Confirm address → Book

**API Call:**

```dart
POST /api/v1/bookings
Body: {
  "waste_type": "Dry Waste",
  "booking_date": "2026-03-20T00:00:00Z",     // ISO 8601
  "time_slot": "6:00 AM - 8:00 AM",
  "address": "123 Main St, Ward 5",
  "latitude": 19.0760,
  "longitude": 72.8777
}
```

**Waste Types:**
- `Dry Waste`
- `Wet Waste`
- `E-Waste`
- `Mixed Waste`
- `Hazardous Waste`

**Time Slots:**
- `6:00 AM - 8:00 AM`
- `8:00 AM - 10:00 AM`
- `10:00 AM - 12:00 PM`
- `2:00 PM - 4:00 PM`
- `4:00 PM - 6:00 PM`

**Frontend Tasks:**
- [ ] Date picker (minimum: tomorrow)
- [ ] Time slot dropdown
- [ ] Waste type selector with icons
- [ ] Address auto-fill from GPS or manual entry
- [ ] Show booking confirmation with estimated points

---

#### 6. My Bookings Screen

```dart
// List bookings
GET /api/v1/bookings?status=pending&page=1&limit=20

// Get single booking
GET /api/v1/bookings/:id
```

**Status flow:** `pending` → `assigned` → `completed` or `cancelled`

**Frontend Tasks:**
- [ ] Show collector info when status is `assigned`
- [ ] Show weight + points earned when `completed`
- [ ] Cancel button for `pending` bookings
- [ ] Real-time status update via polling or push notification

---

#### 7. Wallet Screen

**Shows:** Points balance, wallet balance (₹), redeem button, withdraw button, transaction history

**API Calls:**

```dart
// Get wallet
GET /api/v1/wallet
Response: { "data": { "points": 500, "wallet_balance": 200.00 } }

// Redeem points to wallet
POST /api/v1/wallet/redeem
Body: { "points": 100 }    // converts to ₹100 in wallet
// Error if insufficient points

// Withdraw to bank
POST /api/v1/wallet/withdraw
Body: { "amount": 500.00 }
// Min: ₹100, Max: ₹50,000/day
// Error if insufficient wallet balance

// Transaction history
GET /api/v1/wallet/transactions?type=earned&page=1&limit=20
// type filter: earned, redeemed, withdrawn
```

**Frontend Tasks:**
- [ ] Points → Wallet conversion UI with slider or input
- [ ] "Redeem All" quick button
- [ ] Withdraw flow: enter amount → confirm → show pending status
- [ ] Transaction list with icons per type (earned=green, redeemed=blue, withdrawn=orange)
- [ ] Show "Minimum ₹100" validation message

---

#### 8. Leaderboard Screen

```dart
// Global leaderboard
GET /api/v1/leaderboard?limit=100

// Ward leaderboard
GET /api/v1/leaderboard?ward=Ward5&limit=100

// My rank
GET /api/v1/leaderboard/me
GET /api/v1/leaderboard/me?ward=Ward5
Response: { "data": { "user_id": "...", "score": 500, "rank": 12 } }
```

**Frontend Tasks:**
- [ ] Tab: Global vs My Ward
- [ ] Top 3 highlighted with medal icons
- [ ] Current user highlighted in list
- [ ] Show rank badge at top: "You're #12!"

---

#### 9. Badges / Achievements Screen

```dart
// All badges with user progress
GET /api/v1/badges/user
Response: {
  "data": [
    {
      "badge": { "name": "First Step", "description": "Submit your first report", "tier": "bronze", "trigger_value": 1 },
      "progress": 1,
      "is_unlocked": true,
      "unlocked_at": "2026-03-10T..."
    },
    {
      "badge": { "name": "Reporter", "description": "Submit 10 reports", "tier": "silver", "trigger_value": 10 },
      "progress": 5,
      "is_unlocked": false
    }
  ]
}
```

**15 Badges:**

| Badge | Requirement | Tier | Bonus |
|-------|------------|------|-------|
| First Step | 1 report | Bronze | +5 |
| Reporter | 10 reports | Silver | +20 |
| Clean Crusader | 50 reports | Gold | +100 |
| Check-in Champ | 7-day streak | Bronze | +10 |
| Streak Master | 30-day streak | Gold | +50 |
| Eco Warrior | 100 kg collected | Gold | +100 |
| Point Millionaire | 1000 points | Gold | +50 |
| Community Hero | Top 10 leaderboard | Gold | +100 |
| Speed Reporter | 5 reports in 1 day | Silver | +25 |
| First Pickup | 1 booking | Bronze | +5 |
| Waste Warrior | 10 bookings | Silver | +25 |
| Generous | 1st redemption | Bronze | +5 |
| Big Spender | ₹500 redeemed | Silver | +25 |
| Night Owl | Report 12am-4am | Silver | +15 |
| Top Earner | 500 pts in a week | Gold | +50 |

**Frontend Tasks:**
- [ ] Grid layout with locked/unlocked state
- [ ] Progress bar under each badge (progress / trigger_value)
- [ ] Tier colors: Bronze=#CD7F32, Silver=#C0C0C0, Gold=#FFD700
- [ ] Unlock animation when a new badge is earned

---

#### 10. Profile / Settings Screen

```dart
// Update profile
PATCH /api/v1/user/profile
Body: {
  "name": "Abhay Indwar",
  "ward": "Ward 5",
  "address": "123 Main St, Mumbai",
  "fcm_token": "firebase_token_here"
}
```

**Frontend Tasks:**
- [ ] Edit name, ward, address
- [ ] Send FCM token on login and token refresh
- [ ] Logout → clear stored JWT
- [ ] Show app version, help link

---

## Global Frontend Setup

### API Client (Dio)

```dart
// lib/core/api_client.dart
class ApiClient {
  static const baseUrl = 'https://your-backend-url.com/api/v1';

  late Dio _dio;

  ApiClient() {
    _dio = Dio(BaseOptions(
      baseUrl: baseUrl,
      connectTimeout: Duration(seconds: 10),
      receiveTimeout: Duration(seconds: 10),
      headers: {'Content-Type': 'application/json'},
    ));

    _dio.interceptors.add(InterceptorsWrapper(
      onRequest: (options, handler) async {
        final storage = FlutterSecureStorage();
        final token = await storage.read(key: 'jwt_token');
        if (token != null) {
          options.headers['Authorization'] = 'Bearer $token';
        }
        handler.next(options);
      },
      onError: (error, handler) {
        if (error.response?.statusCode == 401) {
          // Token expired → navigate to login
        }
        handler.next(error);
      },
    ));
  }
}
```

### Response Parsing

Every backend response follows this shape:

```dart
class ApiResponse<T> {
  final bool success;
  final String message;
  final T? data;
  final String? error;
}
```

### FCM Token Registration

```dart
// On app start + on token refresh
final fcmToken = await FirebaseMessaging.instance.getToken();
await dio.patch('/user/profile', data: {'fcm_token': fcmToken});

// Listen for token refresh
FirebaseMessaging.instance.onTokenRefresh.listen((newToken) {
  dio.patch('/user/profile', data: {'fcm_token': newToken});
});
```

### Push Notification Handling

Backend sends these notification types via `data` field:

| data.type | Action |
|-----------|--------|
| `report_resolved` | Navigate to report detail |
| `booking_assigned` | Navigate to booking detail |
| `collector_nearby` | Show ETA overlay |
| `badge_unlocked` | Show badge unlock animation |
| `points_earned` | Show points toast |
| `leaderboard_rank` | Navigate to leaderboard |
| `withdrawal_success` | Navigate to wallet |
| `withdrawal_failed` | Show error + navigate to wallet |
| `new_booking` | (Collector app) Navigate to booking |

---

## Suggested Flutter Folder Structure

```
lib/
├── main.dart
├── core/
│   ├── api_client.dart          # Dio setup + interceptors
│   ├── api_response.dart        # Response model
│   ├── constants.dart           # Base URL, issue types, time slots
│   └── storage.dart             # Secure storage helper
├── models/
│   ├── user.dart
│   ├── report.dart
│   ├── booking.dart
│   ├── transaction.dart
│   ├── badge.dart
│   └── leaderboard_entry.dart
├── services/
│   ├── auth_service.dart        # send-otp, verify-otp
│   ├── user_service.dart        # profile, check-in
│   ├── report_service.dart      # CRUD reports
│   ├── booking_service.dart     # CRUD bookings
│   ├── wallet_service.dart      # wallet, redeem, withdraw, transactions
│   ├── leaderboard_service.dart
│   ├── badge_service.dart
│   └── notification_service.dart # FCM setup
├── providers/                    # State management
│   ├── auth_provider.dart
│   ├── user_provider.dart
│   └── ...
├── screens/
│   ├── login/
│   ├── home/
│   ├── report/
│   ├── booking/
│   ├── wallet/
│   ├── leaderboard/
│   ├── badges/
│   └── profile/
└── widgets/                      # Shared components
    ├── points_badge.dart
    ├── status_chip.dart
    └── ...
```

---

## Deployment Checklist (Before Going Live)

### Backend
- [ ] Deploy to Railway / Render / AWS / DigitalOcean
- [ ] Set all environment variables in production
- [ ] Set `GIN_MODE=release`
- [ ] Use strong `JWT_SECRET` (min 32 chars)
- [ ] Enable PostgreSQL SSL in production (`DB_SSLMODE=require`)
- [ ] Set up database backups

### External Services
- [ ] MSG91 account + template approved for OTP
- [ ] Razorpay account with RazorpayX enabled for payouts
- [ ] Cloudflare R2 bucket created + public access configured
- [ ] Firebase project created + service account JSON downloaded
- [ ] FCM enabled in Firebase console

### Flutter
- [ ] Update `baseUrl` to production backend URL
- [ ] Add `google-services.json` (Android) + `GoogleService-Info.plist` (iOS)
- [ ] Configure deep links for notification taps
- [ ] Add app icon, splash screen
- [ ] Test on real device with real OTP

---

## Common Error Handling

| HTTP Code | Meaning | Flutter Action |
|-----------|---------|----------------|
| 200 | Success | Parse `data` field |
| 201 | Created | Parse `data` field |
| 400 | Bad request | Show `error` message to user |
| 401 | Unauthorized | Clear token, redirect to login |
| 404 | Not found | Show "not found" UI |
| 429 | Rate limited | Show "try again in X minutes" |
| 500 | Server error | Show generic error, retry option |

---

## Quick Reference: All API Endpoints

```
PUBLIC:
  POST   /api/v1/auth/send-otp
  POST   /api/v1/auth/verify-otp

USER (Bearer token required):
  GET    /api/v1/user/profile
  PATCH  /api/v1/user/profile
  POST   /api/v1/user/checkin
  POST   /api/v1/reports                    (multipart/form-data)
  GET    /api/v1/reports                    (?status=&issue_type=&page=&limit=)
  GET    /api/v1/reports/:id
  PATCH  /api/v1/reports/:id/status
  POST   /api/v1/bookings
  GET    /api/v1/bookings                   (?status=&page=&limit=)
  GET    /api/v1/bookings/:id
  PATCH  /api/v1/bookings/:id/status
  GET    /api/v1/wallet
  POST   /api/v1/wallet/redeem
  POST   /api/v1/wallet/withdraw
  POST   /api/v1/payment/verify
  GET    /api/v1/wallet/transactions        (?type=&page=&limit=)
  GET    /api/v1/leaderboard                (?ward=&limit=)
  GET    /api/v1/leaderboard/me             (?ward=)
  GET    /api/v1/badges
  GET    /api/v1/badges/user

COLLECTOR (Collector Bearer token required):
  POST   /api/v1/collector/send-otp
  POST   /api/v1/collector/verify-otp
  GET    /api/v1/collector/profile
  GET    /api/v1/collector/bookings
  PATCH  /api/v1/collector/bookings/:id/complete
  PATCH  /api/v1/collector/location
```
