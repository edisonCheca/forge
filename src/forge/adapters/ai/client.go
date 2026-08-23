package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
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
	if req.IssueID != "" {
		systemPromptBuilder.WriteString(fmt.Sprintf("CRITICAL: You MUST include the issue reference '%s' at the very end of the subject line (e.g., 'feat(scope): subject %s'). ", req.IssueID, req.IssueID))
	}
	systemPromptBuilder.WriteString("Do not include markdown formatting, backticks, or explanatory filler text. Just output the clean commit message.")

	userPrompt := fmt.Sprintf("Here is the git diff of the staged changes:\n\n%s", req.Context.RawDiff)
	if req.ExtraContext != "" {
		userPrompt += fmt.Sprintf("\n\nAdditional developer notes / context to consider:\n%s", req.ExtraContext)
	}

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
		if req.IssueID != "" && !strings.Contains(generatedMessage, req.IssueID) {
			lines := strings.SplitN(generatedMessage, "\n", 2)
			lines[0] = strings.TrimSpace(lines[0]) + " (" + req.IssueID + ")"
			generatedMessage = strings.Join(lines, "\n")
		}

		return &core.CommitProposal{
			Subject:     generatedMessage,
			GeneratedAt: time.Now(),
			ModelUsed:   a.model,
		}, nil
	}

	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

// GeneratePullRequest envía el historial de commits y nombre de rama a la IA para sintetizar un PR completo.
func (a *OpenAIAdapter) GeneratePullRequest(ctx context.Context, req *core.PRGenerateRequest) (*core.PRProposal, error) {
	if a.apiKey == "" {
		return nil, fmt.Errorf("AI API key is missing")
	}

	storyGuidance := ""
	closesText := "Closes #[ID]"
	titleExample := "feat(scope): title"
	if req.StoryID != "" {
		closesText = fmt.Sprintf("Closes %s", req.StoryID)
		if req.StoryCode != "" {
			titleExample = fmt.Sprintf("feat(scope): [%s] title (%s)", req.StoryCode, req.StoryID)
		} else {
			titleExample = fmt.Sprintf("feat(scope): title (%s)", req.StoryID)
		}
		storyGuidance = fmt.Sprintf("\nDOMAIN RULE: The Pull Request always represents the User Story identified by the branch name (User Story ID: '%s', Code: '%s'). The issue tags found inside commit logs represent component subtasks. Therefore, the PR Title suffix and '## Historia de Usuario' MUST reference only the User Story ID from the branch ('%s').", req.StoryID, req.StoryCode, req.StoryID)
	}

	extraGuidance := ""
	if req.ExtraContext != "" {
		extraGuidance = "\n3. Context/Author Notes: Incorporate the provided 'Contexto Adicional / Notas del Autor' naturally within the Pull Request body (e.g., in 'Resumen Ejecutivo' or in a new section 'Decisiones de Diseño')."
	}

	systemPrompt := fmt.Sprintf(`You are a Senior Lead Software Engineer generating a professional GitHub Pull Request proposal.
Output strictly a JSON object with two keys: "title" and "body". Do not include extra wrappers or text outside JSON.
Language requirement: write the entire body in '%s'.
CRITICAL: DO NOT use any emojis anywhere in the title or body. Maintain a clean, professional, minimal, CLI-native tone.%s%s
1. Title: Must follow Conventional Commits format (e.g., '%s').
2. Body format:
## Historia de Usuario
%s

## Resumen Ejecutivo
(Concise 2-3 sentence technical summary of what was built or changed across the branch)

## Subtareas y Cambios Principales
(Bulleted list summarizing the commits. Extract any subtask issue tags like (#10) and format clearly without emojis)

## Plan de Verificación
(Brief bulleted list of verification steps or unit tests executed)`, req.Language, storyGuidance, extraGuidance, titleExample, closesText)

	userPrompt := fmt.Sprintf("Branch Name: %s\nTarget Base Branch: %s\n\nCommit History:\n%s", req.Branch, req.BaseBranch, strings.Join(req.CommitLogs, "\n"))
	if req.ExtraContext != "" {
		userPrompt += fmt.Sprintf("\n\nContexto Adicional / Notas del Autor:\n%s", req.ExtraContext)
	}

	payload := chatRequest{
		Model: a.model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON payload: %w", err)
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
			return nil, err
		}
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+a.apiKey)
		httpReq.Header.Set("HTTP-Referer", "https://github.com/forge/forge")
		httpReq.Header.Set("X-Title", "Forge CLI")

		resp, err := a.httpClient.Do(httpReq)
		if err != nil {
			lastErr = err
			continue
		}

		rawBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= http.StatusInternalServerError || strings.Contains(string(rawBody), "ResourceExhausted") || strings.Contains(string(rawBody), "rate limit") {
			lastErr = fmt.Errorf("status %d: %s", resp.StatusCode, string(rawBody))
			continue
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(rawBody))
		}

		var parsedResp chatResponse
		if err := json.Unmarshal(rawBody, &parsedResp); err != nil {
			return nil, err
		}

		if len(parsedResp.Choices) == 0 || strings.TrimSpace(parsedResp.Choices[0].Message.Content) == "" {
			return nil, fmt.Errorf("empty AI response")
		}

		content := strings.TrimSpace(parsedResp.Choices[0].Message.Content)
		if strings.HasPrefix(content, "```json") {
			content = strings.TrimPrefix(content, "```json")
			content = strings.TrimSuffix(content, "```")
			content = strings.TrimSpace(content)
		} else if strings.HasPrefix(content, "```") {
			content = strings.TrimPrefix(content, "```")
			content = strings.TrimSuffix(content, "```")
			content = strings.TrimSpace(content)
		}

		var prJSON struct {
			Title string `json:"title"`
			Body  string `json:"body"`
		}
		if err := json.Unmarshal([]byte(content), &prJSON); err != nil {
			return nil, fmt.Errorf("failed to decode AI PR JSON (%s): %w", content, err)
		}

		finalTitle := strings.TrimSpace(prJSON.Title)
		finalBody := strings.TrimSpace(prJSON.Body)

		if req.StoryID != "" {
			// 1. Garantía determinista en el título: reemplazar el último (#[0-9]+) por el StoryID o adjuntarlo
			reEndIssue := regexp.MustCompile(`\s*\(#[0-9]+\)\s*$`)
			if reEndIssue.MatchString(finalTitle) {
				finalTitle = reEndIssue.ReplaceAllString(finalTitle, " ("+req.StoryID+")")
			} else if !strings.HasSuffix(finalTitle, "("+req.StoryID+")") {
				finalTitle += " (" + req.StoryID + ")"
			}

			// 2. Garantía determinista en el cuerpo: corregir Closes #[0-9]+ bajo ## Historia de Usuario
			reCloses := regexp.MustCompile(`(?i)(Closes\s+#)[0-9]+`)
			if reCloses.MatchString(finalBody) {
				finalBody = reCloses.ReplaceAllString(finalBody, "${1}"+strings.TrimPrefix(req.StoryID, "#"))
			}
		}

		return &core.PRProposal{
			Title:       finalTitle,
			Body:        finalBody,
			GeneratedAt: time.Now(),
			ModelUsed:   a.model,
		}, nil
	}

	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}
