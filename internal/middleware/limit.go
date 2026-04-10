package middleware

import (
	"github.com/didip/tollbooth/v7"
	"github.com/gin-gonic/gin"
	"net/http"
)

func RateLimit(requestsPerMin float64) gin.HandlerFunc {
	lmt := tollbooth.NewLimiter(requestsPerMin, nil)
	return func(c *gin.Context) {
		httpError := tollbooth.LimitByRequest(lmt, c.Writer, c.Request)
		if httpError != nil {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests. Slow down!"})
			c.Abort()
			return
		}
		c.Next()
	}
}
