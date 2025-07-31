package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/ulule/limiter/v3"
	ginlimiter "github.com/ulule/limiter/v3/drivers/middleware/gin"
	memorystore "github.com/ulule/limiter/v3/drivers/store/memory"
)

func NewRateLimiterMiddleware(v *viper.Viper) gin.HandlerFunc {
	requests := v.GetInt("RATE_LIMIT_REQUESTS")
	window := v.GetInt("RATE_LIMIT_WINDOW")

	if requests <= 0 {
		requests = 1000
	}
	if window <= 0 {
		window = 3600
	}

	rate := limiter.Rate{
		Period: time.Duration(window) * time.Second,
		Limit:  int64(requests),
	}

	store := memorystore.NewStore()

	return ginlimiter.NewMiddleware(limiter.New(store, rate))
}
