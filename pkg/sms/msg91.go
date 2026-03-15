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

// FirebaseAuthService verifies Firebase ID tokens instead of MSG91 OTP.
// OTP send/verify is handled by Flutter's Firebase Phone Auth SDK.
// The backend only verifies the resulting Firebase ID token.
type FirebaseAuthService struct {
	projectID string
	client    *http.Client
}

func NewFirebaseAuthService(projectID string) *FirebaseAuthService {
	return &FirebaseAuthService{
		projectID: projectID,
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

// SendOTP is a no-op — Firebase Phone Auth SDK handles this on the Flutter side.
func (s *FirebaseAuthService) SendOTP(ctx context.Context, phone string) error {
	slog.Info("OTP send handled by Firebase Phone Auth on client side", "phone", phone)
	return nil
}

// VerifyOTP verifies a Firebase ID token (passed as the "otp" field from Flutter).
// Flutter sends the Firebase ID token after successful phone auth.
// Returns nil if valid, error if invalid.
func (s *FirebaseAuthService) VerifyOTP(ctx context.Context, phone, firebaseIDToken string) error {
	// Use the tokeninfo endpoint to validate
	tokenInfoURL := "https://oauth2.googleapis.com/tokeninfo?id_token=" + firebaseIDToken

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, tokenInfoURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("verify token: %w", err)
	}
	defer resp.Body.Close()

	// Alternative: verify via Firebase's secure token endpoint
	if resp.StatusCode != http.StatusOK {
		// Fallback: verify via Firebase Auth REST API
		return s.verifyViaFirebase(ctx, firebaseIDToken, phone)
	}

	return nil
}

func (s *FirebaseAuthService) verifyViaFirebase(ctx context.Context, idToken, expectedPhone string) error {
	url := "https://identitytoolkit.googleapis.com/v1/accounts:lookup"

	payload := fmt.Sprintf(`{"idToken":"%s"}`, idToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Add API key as query param
	q := req.URL.Query()
	q.Add("key", s.projectID)
	req.URL.RawQuery = q.Encode()

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("firebase verify: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		slog.Error("firebase token verification failed", "status", resp.StatusCode, "body", string(body))
		return fmt.Errorf("invalid token")
	}

	var result struct {
		Users []struct {
			PhoneNumber string `json:"phoneNumber"`
		} `json:"users"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	if len(result.Users) == 0 {
		return fmt.Errorf("no user found for token")
	}

	// Verify the phone number matches
	if result.Users[0].PhoneNumber != expectedPhone {
		slog.Error("phone mismatch", "expected", expectedPhone, "got", result.Users[0].PhoneNumber)
		return fmt.Errorf("phone number mismatch")
	}

	slog.Info("Firebase phone auth verified", "phone", expectedPhone)
	return nil
}
