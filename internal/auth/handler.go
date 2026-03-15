package auth

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/indwar7/safaipay-backend/pkg/response"
)

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) SendOTP(c *gin.Context) {
	phone, exists := c.Get("phone_number")
	if !exists {
		response.BadRequest(c, "phone_number is required", "missing phone_number")
		return
	}

	if err := h.service.SendOTP(c.Request.Context(), phone.(string)); err != nil {
		slog.Error("send OTP failed", "error", err)
		response.InternalError(c, "failed to send OTP")
		return
	}

	response.Success(c, "OTP sent successfully", nil)
}

func (h *Handler) VerifyOTP(c *gin.Context) {
	var req OTPVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request", err.Error())
		return
	}

	result, err := h.service.VerifyOTP(c.Request.Context(), req.PhoneNumber, req.OTP)
	if err != nil {
		slog.Error("verify OTP failed", "error", err)
		response.Unauthorized(c, "invalid OTP")
		return
	}

	response.Success(c, "OTP verified successfully", result)
}

func (h *Handler) SendCollectorOTP(c *gin.Context) {
	phone, exists := c.Get("phone_number")
	if !exists {
		response.BadRequest(c, "phone_number is required", "missing phone_number")
		return
	}

	if err := h.service.SendCollectorOTP(c.Request.Context(), phone.(string)); err != nil {
		slog.Error("send collector OTP failed", "error", err)
		response.InternalError(c, "failed to send OTP")
		return
	}

	response.Success(c, "OTP sent successfully", nil)
}

func (h *Handler) VerifyCollectorOTP(c *gin.Context) {
	var req OTPVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request", err.Error())
		return
	}

	result, err := h.service.VerifyCollectorOTP(c.Request.Context(), req.PhoneNumber, req.OTP)
	if err != nil {
		slog.Error("verify collector OTP failed", "error", err)
		response.Unauthorized(c, "invalid OTP")
		return
	}

	response.Success(c, "OTP verified successfully", result)
}
