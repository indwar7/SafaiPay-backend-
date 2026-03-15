package report

import (
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/indwar7/safaipay-backend/pkg/response"
	"github.com/indwar7/safaipay-backend/pkg/storage"
)

type Handler struct {
	service        Service
	storageService *storage.R2Service
}

func NewHandler(s Service, storageSvc *storage.R2Service) *Handler {
	return &Handler{service: s, storageService: storageSvc}
}

func (h *Handler) CreateReport(c *gin.Context) {
	userID := c.GetString("userID")

	var req CreateReportRequest
	req.IssueType = c.PostForm("issue_type")
	req.Description = c.PostForm("description")

	if req.IssueType == "" {
		response.BadRequest(c, "issue_type is required", "missing issue_type")
		return
	}

	latStr := c.PostForm("latitude")
	lngStr := c.PostForm("longitude")
	req.Address = c.PostForm("address")

	if latStr != "" {
		var lat float64
		if _, err := fmt.Sscanf(latStr, "%f", &lat); err == nil {
			req.Latitude = lat
		}
	}
	if lngStr != "" {
		var lng float64
		if _, err := fmt.Sscanf(lngStr, "%f", &lng); err == nil {
			req.Longitude = lng
		}
	}

	var imageURL string
	file, header, err := c.Request.FormFile("image")
	if err == nil {
		defer file.Close()
		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "image/jpeg"
		}
		url, err := h.storageService.Upload(c.Request.Context(), "reports", userID, file, contentType)
		if err != nil {
			slog.Error("image upload failed", "error", err)
			response.InternalError(c, "failed to upload image")
			return
		}
		imageURL = url
	}

	report, err := h.service.CreateReport(c.Request.Context(), userID, &req, imageURL)
	if err != nil {
		slog.Error("create report failed", "error", err)
		response.InternalError(c, "failed to create report")
		return
	}

	response.Created(c, "report created successfully, +5 points earned!", report)
}

func (h *Handler) ListReports(c *gin.Context) {
	var filter ReportFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.BadRequest(c, "invalid filters", err.Error())
		return
	}

	reports, total, err := h.service.ListReports(c.Request.Context(), &filter)
	if err != nil {
		slog.Error("list reports failed", "error", err)
		response.InternalError(c, "failed to list reports")
		return
	}

	response.Success(c, "reports retrieved", gin.H{
		"reports": reports,
		"total":   total,
		"page":    filter.Page,
		"limit":   filter.Limit,
	})
}

func (h *Handler) GetReport(c *gin.Context) {
	id := c.Param("id")

	report, err := h.service.GetReport(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "report not found")
		return
	}

	response.Success(c, "report retrieved", report)
}

func (h *Handler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request", err.Error())
		return
	}

	report, err := h.service.UpdateStatus(c.Request.Context(), id, &req)
	if err != nil {
		slog.Error("update report status failed", "error", err)
		response.InternalError(c, "failed to update status")
		return
	}

	response.Success(c, "status updated", report)
}
