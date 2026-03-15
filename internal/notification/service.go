package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"golang.org/x/oauth2/google"
)

type Service interface {
	SendToUser(ctx context.Context, fcmToken, title, body string, data map[string]string) error
	SendToCollector(ctx context.Context, fcmToken, title, body string, data map[string]string) error
	Send(ctx context.Context, fcmToken string, n *Notification) error
}

type fcmService struct {
	projectID          string
	serviceAccountJSON string
	client             *http.Client
	tokenSource        *cachedTokenSource
}

type cachedTokenSource struct {
	mu          sync.Mutex
	token       string
	expiry      time.Time
	saJSON      string
}

func NewService(projectID, serviceAccountJSON string) Service {
	return &fcmService{
		projectID:          projectID,
		serviceAccountJSON: serviceAccountJSON,
		client:             &http.Client{Timeout: 10 * time.Second},
		tokenSource: &cachedTokenSource{
			saJSON: serviceAccountJSON,
		},
	}
}

func (ts *cachedTokenSource) getToken(ctx context.Context) (string, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.token != "" && time.Now().Before(ts.expiry) {
		return ts.token, nil
	}

	creds, err := google.CredentialsFromJSON(ctx, []byte(ts.saJSON),
		"https://www.googleapis.com/auth/firebase.messaging")
	if err != nil {
		return "", fmt.Errorf("parse service account: %w", err)
	}

	tok, err := creds.TokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("get access token: %w", err)
	}

	ts.token = tok.AccessToken
	ts.expiry = tok.Expiry.Add(-1 * time.Minute)
	return ts.token, nil
}

func (s *fcmService) sendMessage(ctx context.Context, fcmToken, title, body string, data map[string]string) error {
	if fcmToken == "" {
		slog.Warn("empty FCM token, skipping notification")
		return nil
	}

	accessToken, err := s.tokenSource.getToken(ctx)
	if err != nil {
		return fmt.Errorf("get access token: %w", err)
	}

	url := fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", s.projectID)

	message := map[string]interface{}{
		"message": map[string]interface{}{
			"token": fcmToken,
			"notification": map[string]string{
				"title": title,
				"body":  body,
			},
			"data": data,
		},
	}

	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("send FCM message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		slog.Error("FCM send failed", "status", resp.StatusCode, "body", string(respBody))
		return fmt.Errorf("FCM returned status %d", resp.StatusCode)
	}

	slog.Info("FCM notification sent", "title", title)
	return nil
}

func (s *fcmService) SendToUser(ctx context.Context, fcmToken, title, body string, data map[string]string) error {
	return s.sendMessage(ctx, fcmToken, title, body, data)
}

func (s *fcmService) SendToCollector(ctx context.Context, fcmToken, title, body string, data map[string]string) error {
	return s.sendMessage(ctx, fcmToken, title, body, data)
}

func (s *fcmService) Send(ctx context.Context, fcmToken string, n *Notification) error {
	return s.sendMessage(ctx, fcmToken, n.Title, n.Body, n.Data)
}
