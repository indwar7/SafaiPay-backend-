package user

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

func (h *Handler) GetProfile(c *gin.Context) {
	userID := c.GetString("userID")

	profile, err := h.service.GetProfile(c.Request.Context(), userID)
	if err != nil {
		slog.Error("get profile failed", "error", err, "userID", userID)
		response.NotFound(c, "user not found")
		return
	}

	response.Success(c, "profile retrieved", profile)
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	userID := c.GetString("userID")

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request", err.Error())
		return
	}

	profile, err := h.service.UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		slog.Error("update profile failed", "error", err, "userID", userID)
		response.InternalError(c, "failed to update profile")
		return
	}

	response.Success(c, "profile updated", profile)
}

func (h *Handler) DailyCheckIn(c *gin.Context) {
	userID := c.GetString("userID")

	profile, err := h.service.DailyCheckIn(c.Request.Context(), userID)
	if err != nil {
		if err.Error() == "already checked in today" {
			response.BadRequest(c, "already checked in today", err.Error())
			return
		}
		slog.Error("check-in failed", "error", err, "userID", userID)
		response.InternalError(c, "check-in failed")
		return
	}

	response.Success(c, "check-in successful! +2 points", profile)
}
