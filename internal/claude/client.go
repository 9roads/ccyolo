package claude

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
}

type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type Response struct {
	Content []ContentBlock `json:"content"`
	Error   *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type SafetyResult struct {
	Approve bool   `json:"approve"`
	Reason  string `json:"reason"`
}

func EvaluateSafety(apiKey, model, prompt string, toolName string, toolInput map[string]interface{}) (*bool, string, error) {
	inputJSON, _ := json.MarshalIndent(toolInput, "", "  ")

	fullPrompt := fmt.Sprintf(`%s

Tool: %s
Input: %s

Respond with ONLY valid JSON: {"approve": true/false, "reason": "one sentence"}`, prompt, toolName, string(inputJSON))

	reqBody := Request{
		Model:     model,
		MaxTokens: 150,
		Messages: []Message{
			{Role: "user", Content: fullPrompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, "", err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, "", err
	}

	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("content-type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, "", err
	}

	if response.Error != nil {
		return nil, "", fmt.Errorf("API error: %s", response.Error.Message)
	}

	if len(response.Content) == 0 {
		return nil, "", fmt.Errorf("empty response")
	}

	content := response.Content[0].Text

	// Handle markdown code blocks
	if matched, _ := regexp.MatchString("```", content); matched {
		re := regexp.MustCompile("```(?:json)?\\s*(\\{.*?\\})\\s*```")
		matches := re.FindStringSubmatch(content)
		if len(matches) > 1 {
			content = matches[1]
		}
	}

	var result SafetyResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		// Try to extract from text
		if regexp.MustCompile(`"approve"\s*:\s*true`).MatchString(content) {
			result.Approve = true
			result.Reason = "parsed from text"
		} else if regexp.MustCompile(`"approve"\s*:\s*false`).MatchString(content) {
			result.Approve = false
			result.Reason = "parsed from text"
		} else {
			return nil, "", fmt.Errorf("could not parse response: %s", content[:min(100, len(content))])
		}
	}

	return &result.Approve, result.Reason, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ValidateAPIKey tests if an API key is valid by making a minimal API call
func ValidateAPIKey(apiKey string) error {
	reqBody := Request{
		Model:     "claude-haiku-4-5-20251001",
		MaxTokens: 1,
		Messages: []Message{
			{Role: "user", Content: "hi"},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("content-type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return fmt.Errorf("invalid API key")
	}

	if resp.StatusCode == 403 {
		return fmt.Errorf("API key doesn't have permission")
	}

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		var response Response
		if json.Unmarshal(body, &response) == nil && response.Error != nil {
			return fmt.Errorf("%s", response.Error.Message)
		}
		return fmt.Errorf("API error (status %d)", resp.StatusCode)
	}

	return nil
}
