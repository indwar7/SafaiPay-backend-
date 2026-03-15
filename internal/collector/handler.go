package collector

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/indwar7/safaipay-backend/internal/booking"
	"github.com/indwar7/safaipay-backend/internal/notification"
	"github.com/indwar7/safaipay-backend/pkg/response"
	"github.com/indwar7/safaipay-backend/pkg/storage"
)

type Handler struct {
	service        Service
	bookingService booking.Service
	notifService   notification.Service
	storageService *storage.R2Service
}

func NewHandler(s Service, bookingSvc booking.Service, notifSvc notification.Service, storageSvc *storage.R2Service) *Handler {
	return &Handler{
		service:        s,
		bookingService: bookingSvc,
		notifService:   notifSvc,
		storageService: storageSvc,
	}
}

func (h *Handler) GetProfile(c *gin.Context) {
	collectorID := c.GetString("collectorID")

	profile, err := h.service.GetProfile(c.Request.Context(), collectorID)
	if err != nil {
		slog.Error("get collector profile failed", "error", err)
		response.NotFound(c, "collector not found")
		return
	}

	response.Success(c, "profile retrieved", profile)
}

func (h *Handler) GetBookings(c *gin.Context) {
	collectorID := c.GetString("collectorID")

	var filter booking.BookingFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.BadRequest(c, "invalid filters", err.Error())
		return
	}

	bookings, total, err := h.bookingService.GetCollectorBookings(c.Request.Context(), collectorID, &filter)
	if err != nil {
		slog.Error("get collector bookings failed", "error", err)
		response.InternalError(c, "failed to get bookings")
		return
	}

	response.Success(c, "bookings retrieved", gin.H{
		"bookings": bookings,
		"total":    total,
		"page":     filter.Page,
		"limit":    filter.Limit,
	})
}

func (h *Handler) CompleteBooking(c *gin.Context) {
	collectorID := c.GetString("collectorID")
	bookingID := c.Param("id")

	var req CompleteBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request", err.Error())
		return
	}

	if req.Weight <= 0 {
		response.BadRequest(c, "weight must be positive", "invalid weight")
		return
	}

	var imageURL string
	file, header, err := c.Request.FormFile("image")
	if err == nil {
		defer file.Close()
		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "image/jpeg"
		}
		url, err := h.storageService.Upload(c.Request.Context(), "bookings", collectorID, file, contentType)
		if err != nil {
			slog.Error("image upload failed", "error", err)
		} else {
			imageURL = url
		}
	}

	result, err := h.bookingService.CompleteBooking(c.Request.Context(), bookingID, collectorID, req.Weight, imageURL)
	if err != nil {
		slog.Error("complete booking failed", "error", err)
		response.InternalError(c, "failed to complete booking")
		return
	}

	if err := h.service.AddCollected(c.Request.Context(), collectorID, req.Weight); err != nil {
		slog.Error("failed to update collector stats", "error", err)
	}

	response.Success(c, "booking completed", result)
}

func (h *Handler) UpdateLocation(c *gin.Context) {
	collectorID := c.GetString("collectorID")

	var req UpdateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request", err.Error())
		return
	}

	if err := h.service.UpdateLocation(c.Request.Context(), collectorID, req.Latitude, req.Longitude); err != nil {
		slog.Error("update location failed", "error", err)
		response.InternalError(c, "failed to update location")
		return
	}

	response.Success(c, "location updated", nil)
}
