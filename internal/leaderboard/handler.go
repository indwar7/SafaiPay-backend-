package leaderboard

import (
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/indwar7/safaipay-backend/pkg/response"
)

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) GetLeaderboard(c *gin.Context) {
	ward := c.Query("ward")
	limitStr := c.DefaultQuery("limit", "100")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 100
	}

	entries, err := h.service.GetTopN(c.Request.Context(), ward, limit)
	if err != nil {
		slog.Error("get leaderboard failed", "error", err)
		response.InternalError(c, "failed to get leaderboard")
		return
	}

	response.Success(c, "leaderboard retrieved", gin.H{
		"leaderboard": entries,
		"ward":        ward,
		"limit":       limit,
	})
}

func (h *Handler) GetMyRank(c *gin.Context) {
	userID := c.GetString("userID")
	ward := c.Query("ward")

	rank, err := h.service.GetUserRank(c.Request.Context(), userID, ward)
	if err != nil {
		slog.Error("get user rank failed", "error", err)
		response.InternalError(c, "failed to get rank")
		return
	}

	response.Success(c, "rank retrieved", rank)
}
