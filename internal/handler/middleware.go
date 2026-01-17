package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) apiKeyMiddleware(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		headerKey := c.GetHeader("X-API-KEY")
		if headerKey != apiKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
			return
		}
		c.Next()
	}
}
