package handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/krtu0p/code-reviewer/chain"
)

type ReviewRequest struct {
	Language string `json:"language"`
	Code     string `json:"code"`
}

func ReviewHandler(c *gin.Context) {
	var req ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "API key not configured"})
		return
	}

	review, raw, err := chain.RunReviewWithKey(c.Request.Context(), chain.ReviewRequest{
		Language: req.Language,
		Code:     req.Code,
	}, apiKey)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to run review",
			"detail": err.Error(),
			"raw":    raw,
		})
		return
	}

	c.JSON(http.StatusOK, review)
}