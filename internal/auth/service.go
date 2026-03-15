package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/indwar7/safaipay-backend/config"
	"github.com/indwar7/safaipay-backend/internal/collector"
	"github.com/indwar7/safaipay-backend/internal/user"
	"github.com/indwar7/safaipay-backend/pkg/middleware"
	"github.com/indwar7/safaipay-backend/pkg/sms"
)

type Service interface {
	SendOTP(ctx context.Context, phone string) error
	VerifyOTP(ctx context.Context, phone, otp string) (*AuthResponse, error)
	SendCollectorOTP(ctx context.Context, phone string) error
	VerifyCollectorOTP(ctx context.Context, phone, otp string) (*AuthResponse, error)
}

type service struct {
	smsService       *sms.MSG91Service
	userService      user.Service
	collectorService collector.Service
	jwtConfig        *config.JWTConfig
}

func NewService(smsSvc *sms.MSG91Service, userSvc user.Service, collectorSvc collector.Service, jwtCfg *config.JWTConfig) Service {
	return &service{
		smsService:       smsSvc,
		userService:      userSvc,
		collectorService: collectorSvc,
		jwtConfig:        jwtCfg,
	}
}

func (s *service) SendOTP(ctx context.Context, phone string) error {
	return s.smsService.SendOTP(ctx, phone)
}

func (s *service) VerifyOTP(ctx context.Context, phone, otp string) (*AuthResponse, error) {
	if err := s.smsService.VerifyOTP(ctx, phone, otp); err != nil {
		return nil, err
	}

	u, err := s.userService.GetOrCreateByPhone(ctx, phone)
	if err != nil {
		return nil, fmt.Errorf("get or create user: %w", err)
	}

	token, err := s.generateToken(u.ID.String(), phone, "user")
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &AuthResponse{
		Token: token,
		User:  u.ToResponse(),
	}, nil
}

func (s *service) SendCollectorOTP(ctx context.Context, phone string) error {
	return s.smsService.SendOTP(ctx, phone)
}

func (s *service) VerifyCollectorOTP(ctx context.Context, phone, otp string) (*AuthResponse, error) {
	if err := s.smsService.VerifyOTP(ctx, phone, otp); err != nil {
		return nil, err
	}

	c, err := s.collectorService.GetOrCreateByPhone(ctx, phone)
	if err != nil {
		return nil, fmt.Errorf("get or create collector: %w", err)
	}

	token, err := s.generateToken(c.ID.String(), phone, "collector")
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &AuthResponse{
		Token: token,
		User:  c,
	}, nil
}

func (s *service) generateToken(userID, phone, role string) (string, error) {
	claims := middleware.Claims{
		UserID: userID,
		Phone:  phone,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.jwtConfig.ExpiryHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtConfig.Secret))
}
