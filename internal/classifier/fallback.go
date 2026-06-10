package classifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	geminiEndpoint  = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-exp:generateContent"
	fallbackTimeout = 5 * time.Second
)

const geminiSystemPrompt = `You are a model selection advisor for Tokensense. Given a developer task description, classify it.

Output ONLY valid JSON matching this exact schema:
{"task_type": string, "complexity": string, "confidence": number}

task_type must be exactly one of:
  code_generation, code_review, debugging, test_generation,
  documentation, reasoning, data_extraction, summarization, chat

complexity must be exactly one of: low, medium, high

confidence must be a float between 0.0 and 1.0

Rules:
- low: routine, repetitive, template-driven tasks
- medium: requires understanding context or moderate judgment  
- high: synthesis, security analysis, architectural judgment, research

Return ONLY the JSON object. No explanation. No markdown. No prose.`

// GeminiFallback classifies tasks using the Gemini Flash API.
type GeminiFallback struct {
	apiKey     string
	httpClient *http.Client
}

// NewGeminiFallback creates a Gemini Flash fallback classifier.
func NewGeminiFallback(apiKey string) *GeminiFallback {
	return &GeminiFallback{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: fallbackTimeout,
		},
	}
}

type geminiRequest struct {
	SystemInstruction geminiContent    `json:"system_instruction"`
	Contents          []geminiContent  `json:"contents"`
	GenerationConfig  geminiGenConfig  `json:"generationConfig"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenConfig struct {
	Temperature      float64 `json:"temperature"`
	MaxOutputTokens  int     `json:"maxOutputTokens"`
	ResponseMimeType string  `json:"responseMimeType"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// Classify sends the task description to Gemini Flash and parses the response.
func (g *GeminiFallback) Classify(input string) (*ClassificationResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), fallbackTimeout)
	defer cancel()

	reqBody := geminiRequest{
		SystemInstruction: geminiContent{
			Parts: []geminiPart{{Text: geminiSystemPrompt}},
		},
		Contents: []geminiContent{
			{Parts: []geminiPart{{Text: input}}},
		},
		GenerationConfig: geminiGenConfig{
			Temperature:      0.1,
			MaxOutputTokens:  100,
			ResponseMimeType: "application/json",
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal request: %w", err)
	}

	url := geminiEndpoint
	if g.apiKey != "" {
		url += "?key=" + g.apiKey
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var gemResp geminiResponse
	if err := json.Unmarshal(respBody, &gemResp); err != nil {
		return nil, fmt.Errorf("cannot parse gemini response: %w", err)
	}

	if len(gemResp.Candidates) == 0 || len(gemResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from gemini API")
	}

	var result ClassificationResult
	if err := json.Unmarshal([]byte(gemResp.Candidates[0].Content.Parts[0].Text), &result); err != nil {
		return nil, fmt.Errorf("cannot parse classification result: %w", err)
	}

	result.Source = SourceAPIFallback
	return &result, nil
}
