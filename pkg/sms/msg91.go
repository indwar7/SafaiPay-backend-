package sms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/indwar7/safaipay-backend/config"
)

type MSG91Service struct {
	authKey    string
	templateID string
	rdb        *redis.Client
	client     *http.Client
}

func NewMSG91Service(cfg *config.MSG91Config, rdb *redis.Client) *MSG91Service {
	return &MSG91Service{
		authKey:    cfg.AuthKey,
		templateID: cfg.TemplateID,
		rdb:        rdb,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *MSG91Service) SendOTP(ctx context.Context, phone string) error {
	url := "https://control.msg91.com/api/v5/otp"

	payload := map[string]string{
		"template_id": s.templateID,
		"mobile":      phone,
		"authkey":     s.authKey,
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("send OTP: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		slog.Error("MSG91 OTP send failed", "status", resp.StatusCode, "body", string(respBody))
		return fmt.Errorf("MSG91 returned status %d", resp.StatusCode)
	}

	slog.Info("OTP sent", "phone", phone)
	return nil
}

func (s *MSG91Service) VerifyOTP(ctx context.Context, phone, otp string) error {
	attemptsKey := fmt.Sprintf("otp_attempts:%s", phone)

	attempts, _ := s.rdb.Get(ctx, attemptsKey).Int()
	if attempts >= 3 {
		return fmt.Errorf("maximum verification attempts exceeded")
	}

	url := fmt.Sprintf("https://control.msg91.com/api/v5/otp/verify?mobile=%s&otp=%s&authkey=%s",
		phone, otp, s.authKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("verify OTP: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	if result.Type != "success" {
		s.rdb.Incr(ctx, attemptsKey)
		s.rdb.Expire(ctx, attemptsKey, 5*time.Minute)
		return fmt.Errorf("invalid OTP")
	}

	s.rdb.Del(ctx, attemptsKey)
	return nil
}
