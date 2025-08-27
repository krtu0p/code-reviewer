package chain

import (
    "context"
    "encoding/json"
    "os"
    "regexp"
    "strings"

    "github.com/cloudwego/eino/schema"
    "github.com/cloudwego/eino/compose"
    "github.com/cloudwego/eino-ext/components/model/deepseek"
)


type ReviewJSON struct {
  Summary     string   `json:"summary"`
  Issues      []string `json:"issues"`
  Suggestions []string `json:"suggestions"`
  Complexity  string   `json:"complexity"` // low | medium | high
  Score       int      `json:"score"`
}

type ReviewRequest struct {
  Language string `json:"language" binding:"required"`
  Code     string `json:"code" binding:"required"`
}

func BuildReviewChain(ctx context.Context) (*compose.Chain[map[string]any, *schema.Message], error) {
  chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
    APIKey: os.Getenv("DEEPSEEK_API_KEY"),
    Model:  "deepseek-coder",
  })
  if err != nil {
    return nil, err
  }

  chainBuilder := compose.NewChain[map[string]any, *schema.Message]()
  chainBuilder.AppendChatModel(chatModel)

  return chainBuilder, nil
}

func RunReview(ctx context.Context, req ReviewRequest) (*ReviewJSON, string, error) {
  chain, err := BuildReviewChain(ctx)
  if err != nil {
    return nil, "", err
  }

  // Prepareprompt inside map
  inputMap := map[string]any{
    "system": schema.SystemMessage(`You are a strict senior code reviewer...
Return ONLY valid minified JSON with this schema:
{"summary":..., "issues":..., "suggestions":..., "complexity":"low|medium|high", "score":number}
No prose, no markdown, no backticks.`),
    "user": schema.UserMessage("Language: " + req.Language + "\n\nCODE:\n" + req.Code),
  }

  runner, err := chain.Compile(ctx)
  if err != nil {
    return nil, "", err
  }

  outMsg, err := runner.Invoke(ctx, inputMap)
  if err != nil {
    return nil, "", err
  }

  raw := strings.TrimSpace(outMsg.Content)

  // Parse JSON
  parsed := ReviewJSON{}
  if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
    re := regexp.MustCompile(`(?s)\{.*\}`)
    if m := re.FindString(raw); m != "" {
      if err2 := json.Unmarshal([]byte(m), &parsed); err2 == nil {
        return &parsed, m, nil
      }
    }
    return nil, raw, nil
  }

  return &parsed, raw, nil
}
