// handlers/review.go
package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"./chain"
)

func ReviewHandler(c *gin.Context) {
	var req chain.ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request payload",
		})
		return
	}

	ctx := context.Background()

	parsed, raw, err := chain.RunReview(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "failed to run review",
			"detail": err.Error(),
		})
		return
	}

	if parsed != nil {
		// hasil JSON sudah berhasil di-parse ke ReviewJSON
		c.JSON(http.StatusOK, parsed)
	} else {
		// fallback kalau parse gagal → return raw string dari model
		c.JSON(http.StatusOK, gin.H{"raw": raw})
	}
}
