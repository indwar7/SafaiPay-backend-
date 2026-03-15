package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/indwar7/safaipay-backend/pkg/response"
)

func RateLimitOTP(rdb *redis.Client, maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body struct {
			PhoneNumber string `json:"phone_number"`
		}
		if err := c.ShouldBindJSON(&body); err != nil || body.PhoneNumber == "" {
			response.BadRequest(c, "phone_number is required", "invalid request body")
			c.Abort()
			return
		}

		// Re-set body for downstream handlers
		c.Set("phone_number", body.PhoneNumber)

		key := fmt.Sprintf("ratelimit:otp:%s", body.PhoneNumber)
		ctx := context.Background()

		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			response.InternalError(c, "rate limit check failed")
			c.Abort()
			return
		}

		if count == 1 {
			rdb.Expire(ctx, key, window)
		}

		if count > int64(maxRequests) {
			response.Error(c, 429, "too many OTP requests", "please try again later")
			c.Abort()
			return
		}

		c.Next()
	}
}
