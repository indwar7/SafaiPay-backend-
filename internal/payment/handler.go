package payment

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

func (h *Handler) GetWallet(c *gin.Context) {
	userID := c.GetString("userID")

	wallet, err := h.service.GetWallet(c.Request.Context(), userID)
	if err != nil {
		slog.Error("get wallet failed", "error", err)
		response.InternalError(c, "failed to get wallet")
		return
	}

	response.Success(c, "wallet retrieved", wallet)
}

func (h *Handler) RedeemPoints(c *gin.Context) {
	userID := c.GetString("userID")

	var req RedeemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request", err.Error())
		return
	}

	tx, err := h.service.RedeemPoints(c.Request.Context(), userID, &req)
	if err != nil {
		slog.Error("redeem points failed", "error", err)
		response.BadRequest(c, "redemption failed", err.Error())
		return
	}

	response.Success(c, "points redeemed successfully", tx)
}

func (h *Handler) Withdraw(c *gin.Context) {
	userID := c.GetString("userID")

	var req WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request", err.Error())
		return
	}

	if req.Amount < 100 {
		response.BadRequest(c, "minimum withdrawal is ₹100", "amount too low")
		return
	}
	if req.Amount > 50000 {
		response.BadRequest(c, "maximum withdrawal is ₹50,000/day", "amount too high")
		return
	}

	tx, err := h.service.Withdraw(c.Request.Context(), userID, &req)
	if err != nil {
		slog.Error("withdrawal failed", "error", err)
		response.BadRequest(c, "withdrawal failed", err.Error())
		return
	}

	response.Success(c, "withdrawal initiated", tx)
}

func (h *Handler) VerifyPayment(c *gin.Context) {
	var req RazorpayVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request", err.Error())
		return
	}

	if err := h.service.VerifyRazorpay(c.Request.Context(), &req); err != nil {
		response.BadRequest(c, "payment verification failed", err.Error())
		return
	}

	response.Success(c, "payment verified successfully", nil)
}

func (h *Handler) GetTransactions(c *gin.Context) {
	userID := c.GetString("userID")

	var filter TransactionFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.BadRequest(c, "invalid filters", err.Error())
		return
	}

	transactions, total, err := h.service.GetTransactions(c.Request.Context(), userID, &filter)
	if err != nil {
		slog.Error("get transactions failed", "error", err)
		response.InternalError(c, "failed to get transactions")
		return
	}

	response.Success(c, "transactions retrieved", gin.H{
		"transactions": transactions,
		"total":        total,
		"page":         filter.Page,
		"limit":        filter.Limit,
	})
}
