package badges

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

func (h *Handler) GetAllBadges(c *gin.Context) {
	badges, err := h.service.GetAllBadges(c.Request.Context())
	if err != nil {
		slog.Error("get badges failed", "error", err)
		response.InternalError(c, "failed to get badges")
		return
	}

	response.Success(c, "badges retrieved", badges)
}

func (h *Handler) GetUserBadges(c *gin.Context) {
	userID := c.GetString("userID")

	badges, err := h.service.GetUserBadges(c.Request.Context(), userID)
	if err != nil {
		slog.Error("get user badges failed", "error", err)
		response.InternalError(c, "failed to get user badges")
		return
	}

	response.Success(c, "user badges retrieved", badges)
}
