package chain

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

type ReviewJSON struct {
	Summary     string   `json:"summary"`
	Issues      []string `json:"issues"`
	Suggestions []string `json:"suggestions"`
	Complexity  string   `json:"complexity"`
	Score       int      `json:"score"`
}

type ReviewRequest struct {
	Language string `json:"language"`
	Code     string `json:"code"`
}

const (
	openRouterURL    = "https://openrouter.ai/api/v1/chat/completions"
	contentTypeJSON  = "application/json"
	refererHeader    = "https://your-domain.com"
	titleHeader      = "Code Reviewer"
	defaultMaxTokens = 500
	defaultTemp      = 0.1
)

var (
	jsonRegex = regexp.MustCompile(`\{[\s\S]*\}`)
)

func RunReviewWithKey(ctx context.Context, req ReviewRequest, apiKey string) (*ReviewJSON, string, error) {
	if apiKey == "" {
		return nil, "", errors.New("API key is empty")
	}

	payload, err := createPayload(req)
	if err != nil {
		return nil, "", fmt.Errorf("create payload: %w", err)
	}

	respBody, err := makeAPIRequest(ctx, apiKey, payload)
	if err != nil {
		return nil, "", fmt.Errorf("API request: %w", err)
	}

	responseText, err := parseAPIResponse(respBody)
	if err != nil {
		return nil, string(respBody), fmt.Errorf("parse API response: %w", err)
	}

	jsonStr, err := extractJSON(responseText)
	if err != nil {
		return nil, responseText, fmt.Errorf("extract JSON: %w", err)
	}

	parsed, err := parseReviewJSON(jsonStr)
	if err != nil {
		return nil, jsonStr, fmt.Errorf("parse review JSON: %w", err)
	}

	return parsed, jsonStr, nil
}

func createPayload(req ReviewRequest) ([]byte, error) {
	systemPrompt := `You are a strict senior code reviewer. Return ONLY valid JSON with this exact structure:
{
  "summary": "brief summary of the code review",
  "issues": ["list of specific issues found"],
  "suggestions": ["list of improvement suggestions"],
  "complexity": "simple|medium|complex",
  "score": 0-100
}

IMPORTANT: Return ONLY the JSON object. No additional text, no explanations, no reasoning, no markdown formatting.`

	payload := map[string]interface{}{
		"model":       "openai/gpt-4o",
		"max_tokens":  defaultMaxTokens,
		"temperature": defaultTemp,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": systemPrompt,
			},
			{
				"role":    "user",
				"content": fmt.Sprintf("Language: %s\n\nCODE:\n%s", req.Language, req.Code),
			},
		},
	}

	return json.Marshal(payload)
}

func makeAPIRequest(ctx context.Context, apiKey string, payload []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openRouterURL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	setHeaders(req, apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

func setHeaders(req *http.Request, apiKey string) {
	req.Header.Set("Content-Type", contentTypeJSON)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("HTTP-Referer", refererHeader)
	req.Header.Set("X-Title", titleHeader)
}

func parseAPIResponse(respBody []byte) (string, error) {
	var orResp struct {
		Choices []struct {
			Message struct {
				Content   string `json:"content"`
				Reasoning string `json:"reasoning"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(respBody, &orResp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	if orResp.Error.Message != "" {
		return "", fmt.Errorf("API error: %s", orResp.Error.Message)
	}

	if len(orResp.Choices) == 0 {
		return "", errors.New("no choices in response")
	}

	choice := orResp.Choices[0].Message
	if choice.Content == "" && choice.Reasoning == "" {
		return "", errors.New("empty response from AI")
	}

	if choice.Content != "" {
		return choice.Content, nil
	}
	return choice.Reasoning, nil
}

func extractJSON(text string) (string, error) {
	cleanText := strings.TrimSpace(text)

	// Check if entire text is valid JSON
	if isValidJSON(cleanText) {
		return cleanText, nil
	}

	// Remove markdown code blocks
	cleanText = removeMarkdown(cleanText)
	if isValidJSON(cleanText) {
		return cleanText, nil
	}

	// Try regex extraction
	if jsonStr := extractJSONWithRegex(cleanText); jsonStr != "" {
		return jsonStr, nil
	}

	// Try brace matching extraction
	if jsonStr := extractJSONWithBraces(cleanText); jsonStr != "" {
		return jsonStr, nil
	}

	return "", errors.New("no valid JSON found in response")
}

func removeMarkdown(text string) string {
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	return strings.TrimSpace(text)
}

func extractJSONWithRegex(text string) string {
	matches := jsonRegex.FindStringSubmatch(text)
	if len(matches) > 0 && isValidJSON(matches[0]) {
		return matches[0]
	}
	return ""
}

func extractJSONWithBraces(text string) string {
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start != -1 && end != -1 && end > start {
		potentialJSON := text[start : end+1]
		if isValidJSON(potentialJSON) {
			return potentialJSON
		}
	}
	return ""
}

func isValidJSON(s string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(s), &js) == nil
}

func parseReviewJSON(jsonStr string) (*ReviewJSON, error) {
	var parsed ReviewJSON
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return &parsed, nil
}