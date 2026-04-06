package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	// defaultModel is the default OpenRouter model used to generate commit messages.
	defaultModel   = "qwen/qwen3.6-plus:free"
	fallbackModel  = "minimax/minimax-m2.5:free"
	defaultBaseURL = "https://openrouter.ai/api/v1"
)

// openRouterRequest represents the request payload for OpenRouter's chat completions API.
type openRouterRequest struct {
	Model       string              `json:"model"`
	Messages    []openRouterMessage `json:"messages"`
	Temperature float64             `json:"temperature,omitempty"`
}

type openRouterMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// openRouterResponse represents the minimal subset of the OpenRouter response we care about.
type openRouterResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func uniqueModels(models []string) []string {
	seen := make(map[string]struct{}, len(models))
	out := make([]string, 0, len(models))
	for _, model := range models {
		if model == "" {
			continue
		}
		if _, ok := seen[model]; ok {
			continue
		}
		seen[model] = struct{}{}
		out = append(out, model)
	}
	return out
}

func callOpenRouter(apiKey, baseURL, model, prompt string) (string, error) {
	reqBody := openRouterRequest{
		Model: model,
		Messages: []openRouterMessage{
			{
				Role:    "system",
				Content: "You are a helpful assistant that writes semantic git commit messages.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.2,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/chat/completions", strings.TrimRight(baseURL, "/"))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("model %q failed with status %d: %s", model, resp.StatusCode, string(body))
	}

	var result openRouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Choices) == 0 || strings.TrimSpace(result.Choices[0].Message.Content) == "" {
		return "", fmt.Errorf("model %q returned empty response", model)
	}

	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}

// GenerateSubject generates the commit subject.
// If userDescription is non-empty, it is used as additional context alongside the diff.
func GenerateSubject(diff string, commitType string, emoji string, userDescription string) (string, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENROUTER_API_KEY env var is not set")
	}

	baseURL := os.Getenv("OPENROUTER_API_BASE_URL")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	model := os.Getenv("OPENROUTER_MODEL")
	if model == "" {
		model = defaultModel
	}
	modelsToTry := uniqueModels([]string{model, fallbackModel})

	descSection := ""
	if userDescription != "" {
		descSection = fmt.Sprintf("\nUser description: %s\n", userDescription)
	}

	prompt := fmt.Sprintf(`
You are a git commit message generator.
Task: Generate a commit message based on the provided git diff.%s
Result must STRICTLY follow these rules:
1. Format:
<subject>

<body>

2. The <subject> must be in present tense, max 50 chars, no trailing period.
3. The <body> should be 2-3 short sentences explaining *why* the change was made, based on the diff.
4. Do NOT include the commit type ("%s") or emoji ("%s") in the output. I will prepend them to the subject myself.

Diff:
%s
`, descSection, commitType, emoji, diff)

	var subject string
	var lastErr error
	for _, modelName := range modelsToTry {
		subject, lastErr = callOpenRouter(apiKey, baseURL, modelName, prompt)
		if lastErr == nil {
			break
		}
	}
	if lastErr != nil {
		return "", fmt.Errorf("all models failed (%s): %w", strings.Join(modelsToTry, ", "), lastErr)
	}

	// Startup cleaning
	subject = strings.Trim(subject, "\"")
	subject = strings.Trim(subject, "'")
	return subject, nil
}
