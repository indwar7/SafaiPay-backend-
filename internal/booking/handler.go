package booking

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

func (h *Handler) CreateBooking(c *gin.Context) {
	userID := c.GetString("userID")

	var req CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request", err.Error())
		return
	}

	booking, err := h.service.CreateBooking(c.Request.Context(), userID, &req)
	if err != nil {
		slog.Error("create booking failed", "error", err)
		response.InternalError(c, "failed to create booking")
		return
	}

	response.Created(c, "booking created successfully", booking)
}

func (h *Handler) ListBookings(c *gin.Context) {
	userID := c.GetString("userID")

	var filter BookingFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.BadRequest(c, "invalid filters", err.Error())
		return
	}

	bookings, total, err := h.service.GetUserBookings(c.Request.Context(), userID, &filter)
	if err != nil {
		slog.Error("list bookings failed", "error", err)
		response.InternalError(c, "failed to list bookings")
		return
	}

	response.Success(c, "bookings retrieved", gin.H{
		"bookings": bookings,
		"total":    total,
		"page":     filter.Page,
		"limit":    filter.Limit,
	})
}

func (h *Handler) GetBooking(c *gin.Context) {
	id := c.Param("id")

	booking, err := h.service.GetBooking(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "booking not found")
		return
	}

	response.Success(c, "booking retrieved", booking)
}

func (h *Handler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")

	var req UpdateBookingStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request", err.Error())
		return
	}

	booking, err := h.service.UpdateStatus(c.Request.Context(), id, &req)
	if err != nil {
		slog.Error("update booking status failed", "error", err)
		response.InternalError(c, "failed to update status")
		return
	}

	response.Success(c, "booking status updated", booking)
}
