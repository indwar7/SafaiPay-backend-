package sms

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// FirebaseAuthService verifies Firebase ID tokens.
// OTP send/verify is handled by Flutter's Firebase Phone Auth SDK.
// The backend verifies the resulting Firebase ID token and issues its own JWT.
type FirebaseAuthService struct {
	apiKey string // Firebase Web API Key (AIzaSy...)
	client *http.Client
}

func NewFirebaseAuthService(apiKey string) *FirebaseAuthService {
	return &FirebaseAuthService{
		apiKey: apiKey,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// SendOTP is a no-op — Firebase Phone Auth SDK handles this on the Flutter side.
func (s *FirebaseAuthService) SendOTP(ctx context.Context, phone string) error {
	slog.Info("OTP send handled by Firebase Phone Auth on client side", "phone", phone)
	return nil
}

// VerifyOTP verifies a Firebase ID token (passed as the "otp" field from Flutter).
// Uses Firebase Auth REST API: accounts:lookup with the ID token.
func (s *FirebaseAuthService) VerifyOTP(ctx context.Context, phone, firebaseIDToken string) error {
	if firebaseIDToken == "" {
		return fmt.Errorf("firebase token is empty")
	}

	// Use Firebase Auth REST API to look up the user by ID token
	url := fmt.Sprintf("https://identitytoolkit.googleapis.com/v1/accounts:lookup?key=%s", s.apiKey)

	payload := fmt.Sprintf(`{"idToken":"%s"}`, firebaseIDToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("firebase verify: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		slog.Error("firebase token verification failed", "status", resp.StatusCode, "body", string(body))
		return fmt.Errorf("invalid or expired token")
	}

	var result struct {
		Users []struct {
			LocalID     string `json:"localId"`
			PhoneNumber string `json:"phoneNumber"`
		} `json:"users"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	if len(result.Users) == 0 {
		return fmt.Errorf("no user found for this token")
	}

	// Verify the phone number matches
	firebasePhone := result.Users[0].PhoneNumber
	if firebasePhone != phone {
		slog.Error("phone mismatch", "expected", phone, "got", firebasePhone)
		return fmt.Errorf("phone number mismatch")
	}

	slog.Info("Firebase phone auth verified", "phone", phone, "uid", result.Users[0].LocalID)
	return nil
}
