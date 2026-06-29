package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/forge/forge/src/forge/core"
)

// Verificación en tiempo de compilación para asegurar que OpenAIAdapter implementa core.AIPort.
var _ core.AIPort = (*OpenAIAdapter)(nil)

// OpenAIAdapter implementa la comunicación HTTP nativa con cualquier API compatible con OpenAI.
type OpenAIAdapter struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
}

// NewOpenAIAdapter crea una nueva instancia del adaptador de IA agnóstico al proveedor.
func NewOpenAIAdapter(apiKey, baseURL, model string) *OpenAIAdapter {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1/chat/completions"
	}
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &OpenAIAdapter{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatResponse struct {
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// GenerateCommit envía el contexto del repositorio al endpoint configurado y retorna la propuesta de commit.
func (a *OpenAIAdapter) GenerateCommit(ctx context.Context, req *core.GenerateRequest) (*core.CommitProposal, error) {
	if a.apiKey == "" {
		return nil, fmt.Errorf("AI API key is missing or empty")
	}
	if req == nil || req.Context == nil {
		return nil, fmt.Errorf("generate request or repository context is nil")
	}

	// Construcción estricta del System Prompt
	var systemPromptBuilder strings.Builder
	systemPromptBuilder.WriteString("You are an expert software engineer generating a git commit message. ")
	systemPromptBuilder.WriteString(fmt.Sprintf("Output the entire commit message exclusively in the language: '%s'. ", req.Language))

	if req.ConventionalCommit {
		maxLen := req.MaxLength
		if maxLen <= 0 {
			maxLen = 72
		}
		systemPromptBuilder.WriteString("Follow the Conventional Commits format strictly. ")
		systemPromptBuilder.WriteString("You MUST include a scope in parentheses describing the affected component or file (e.g., feat(auth):, docs(readme):, fix(api):). ")
		systemPromptBuilder.WriteString("You MUST include a space after the colon (e.g., 'feat(ui): add' is correct, 'feat(ui):add' is WRONG). ")
		systemPromptBuilder.WriteString(fmt.Sprintf("The subject line must not end with a period and must be under %d characters. ", maxLen))
		systemPromptBuilder.WriteString("Separate the subject from the body with a single blank line. ")
		systemPromptBuilder.WriteString("CRITICAL: For the body, NEVER write long paragraphs. You MUST use a bulleted list (using '-' as bullets). Keep every single bullet point extremely short and concise to guarantee NO LINE exceeds 80 characters. ")
	} else if req.MaxLength > 0 {
		systemPromptBuilder.WriteString(fmt.Sprintf("Keep the subject line under %d characters. ", req.MaxLength))
	}
	systemPromptBuilder.WriteString("Do not include markdown formatting, backticks, or explanatory filler text. Just output the clean commit message.")

	userPrompt := fmt.Sprintf("Here is the git diff of the staged changes:\n\n%s", req.Context.RawDiff)

	payload := chatRequest{
		Model: a.model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPromptBuilder.String()},
			{Role: "user", Content: userPrompt},
		},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON payload for AI request: %w", err)
	}

	maxRetries := 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			backoff := time.Duration(attempt) * 1500 * time.Millisecond
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL, bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to create HTTP request: %w", err)
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+a.apiKey)
		httpReq.Header.Set("HTTP-Referer", "https://github.com/forge/forge")
		httpReq.Header.Set("X-Title", "Forge CLI")

		resp, err := a.httpClient.Do(httpReq)
		if err != nil {
			lastErr = fmt.Errorf("failed to execute HTTP request to AI provider (%s): %w", a.baseURL, err)
			continue
		}

		rawBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read AI response body: %w", err)
			continue
		}

		// Reintentar automáticamente ante saturación de trabajadores (ResourceExhausted), rate limits (429) o errores de servidor (>= 500)
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= http.StatusInternalServerError || strings.Contains(string(rawBody), "ResourceExhausted") || strings.Contains(string(rawBody), "rate limit") {
			lastErr = fmt.Errorf("AI API request to %s failed with status code %d: %s", a.baseURL, resp.StatusCode, strings.TrimSpace(string(rawBody)))
			continue
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("AI API request to %s failed with status code %d: %s", a.baseURL, resp.StatusCode, strings.TrimSpace(string(rawBody)))
		}

		var parsedResp chatResponse
		if err := json.Unmarshal(rawBody, &parsedResp); err != nil {
			return nil, fmt.Errorf("failed to decode AI JSON response: %w. Raw body: %s", err, string(rawBody))
		}

		if parsedResp.Error != nil && parsedResp.Error.Message != "" {
			return nil, fmt.Errorf("AI API returned error: %s", parsedResp.Error.Message)
		}

		if len(parsedResp.Choices) == 0 || strings.TrimSpace(parsedResp.Choices[0].Message.Content) == "" {
			return nil, fmt.Errorf("AI API returned empty response choices. Raw response: %s", strings.TrimSpace(string(rawBody)))
		}

		generatedMessage := strings.TrimSpace(parsedResp.Choices[0].Message.Content)

		return &core.CommitProposal{
			Subject:     generatedMessage,
			GeneratedAt: time.Now(),
			ModelUsed:   a.model,
		}, nil
	}

	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}
