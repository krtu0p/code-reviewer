// handlers/review.go
package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/krtu0p/code-reviewer/chain"
)

func ReviewHandler(c *gin.Context) {
	var req chain.ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// ✅ Read API key from Authorization header
	apiKey := c.GetHeader("Authorization")
	if apiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "API key missing"})
		return
	}
	apiKey = strings.TrimPrefix(apiKey, "Bearer ")

	ctx := context.Background()
	parsed, raw, err := chain.RunReviewWithKey(ctx, req, apiKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "failed to run review",
			"detail": err.Error(),
			"raw":    raw,
		})
		return
	}

	if parsed != nil {
		c.JSON(http.StatusOK, parsed)
	} else {
		c.JSON(http.StatusOK, gin.H{"raw": raw})
	}
}
