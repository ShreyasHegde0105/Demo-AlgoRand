package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"procure-ai/models"
)

const defaultGeminiModel = "gemini-2.5-flash-lite"

type GeminiParserService struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

type geminiGenerateContentRequest struct {
	Contents         []geminiContent        `json:"contents"`
	GenerationConfig geminiGenerationConfig `json:"generationConfig"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerationConfig struct {
	ResponseMIMEType string         `json:"responseMimeType"`
	ResponseSchema   map[string]any `json:"responseSchema"`
	Temperature      float64        `json:"temperature"`
}

type geminiGenerateContentResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func NewGeminiParserServiceFromEnv() *GeminiParserService {
	apiKey := strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))
	model := strings.TrimSpace(os.Getenv("GEMINI_MODEL"))
	if model == "" {
		model = defaultGeminiModel
	}

	timeout := 15 * time.Second
	if raw := strings.TrimSpace(os.Getenv("GEMINI_TIMEOUT_SECONDS")); raw != "" {
		if seconds, err := time.ParseDuration(raw + "s"); err == nil && seconds > 0 {
			timeout = seconds
		}
	}

	return &GeminiParserService{
		apiKey:  apiKey,
		model:   model,
		baseURL: "https://generativelanguage.googleapis.com/v1beta",
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (s *GeminiParserService) Enabled() bool {
	return s != nil && s.apiKey != ""
}

func (s *GeminiParserService) ParseProcurementPrompt(prompt string, topN int) (*models.ParsedProcurementRequest, *models.ProcurementRequest, error) {
	if strings.TrimSpace(prompt) == "" {
		return nil, nil, fmt.Errorf("prompt is required")
	}
	if !s.Enabled() {
		return nil, nil, fmt.Errorf("GEMINI_API_KEY is not configured")
	}

	payload := geminiGenerateContentRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{Text: buildPrompt(prompt)},
				},
			},
		},
		GenerationConfig: geminiGenerationConfig{
			ResponseMIMEType: "application/json",
			ResponseSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"category": map[string]any{
						"type": "string",
					},
					"quantity": map[string]any{
						"type": "integer",
					},
					"unitBudget": map[string]any{
						"type": "number",
					},
					"totalBudget": map[string]any{
						"type": "number",
					},
					"maxDeliveryDays": map[string]any{
						"type": "integer",
					},
					"preferredCities": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "string",
						},
					},
				},
				"required": []string{"category", "quantity"},
			},
			Temperature: 0.1,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", s.baseURL, s.model, s.apiKey)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	var parsedResponse geminiGenerateContentResponse
	if err := json.Unmarshal(respBody, &parsedResponse); err != nil {
		return nil, nil, fmt.Errorf("failed to decode Gemini response: %w", err)
	}
	if parsedResponse.Error != nil {
		return nil, nil, fmt.Errorf("Gemini API error: %s", parsedResponse.Error.Message)
	}
	if resp.StatusCode >= 400 {
		return nil, nil, fmt.Errorf("Gemini API returned status %d", resp.StatusCode)
	}
	if len(parsedResponse.Candidates) == 0 || len(parsedResponse.Candidates[0].Content.Parts) == 0 {
		return nil, nil, fmt.Errorf("Gemini returned no content")
	}

	var parsed models.ParsedProcurementRequest
	if err := json.Unmarshal([]byte(parsedResponse.Candidates[0].Content.Parts[0].Text), &parsed); err != nil {
		return nil, nil, fmt.Errorf("failed to decode parsed procurement JSON: %w", err)
	}

	request, err := normalizeParsedProcurementRequest(parsed, topN)
	if err != nil {
		return nil, nil, err
	}

	return &parsed, request, nil
}

func buildPrompt(prompt string) string {
	return strings.TrimSpace(fmt.Sprintf(`
Extract structured procurement requirements from the user message.

Return JSON only.
Do not include markdown.
If the user says a per-unit budget like "40000 each", set unitBudget to that number and totalBudget to quantity * unitBudget.
If delivery is described like "within a week", convert it to maxDeliveryDays = 7.
If cities are not mentioned, return an empty array for preferredCities.
Normalize category names to concise lowercase labels such as laptops, electronics, medical, raw_materials.

User message:
%s
`, prompt))
}

func normalizeParsedProcurementRequest(parsed models.ParsedProcurementRequest, topN int) (*models.ProcurementRequest, error) {
	category := normalizeCategory(parsed.Category)
	if category == "" {
		return nil, fmt.Errorf("could not determine category from prompt")
	}
	if parsed.Quantity <= 0 {
		return nil, fmt.Errorf("could not determine a valid quantity from prompt")
	}

	budget := parsed.TotalBudget
	if budget <= 0 && parsed.UnitBudget > 0 {
		budget = parsed.UnitBudget * float64(parsed.Quantity)
	}

	request := &models.ProcurementRequest{
		Category:        category,
		Quantity:        parsed.Quantity,
		Budget:          budget,
		MaxDeliveryDays: parsed.MaxDeliveryDays,
		PreferredCities: cleanCities(parsed.PreferredCities),
		TopN:            topN,
	}

	return request, nil
}

func normalizeCategory(value string) string {
	category := strings.TrimSpace(strings.ToLower(value))
	switch category {
	case "laptop", "laptops", "computer", "computers", "electronics", "electronic", "hardware", "it equipment":
		return "electronics"
	case "medical", "medical supplies", "medical equipment", "healthcare":
		return "medical"
	case "raw material", "raw materials", "raw_material", "raw_materials", "industrial materials":
		return "raw_materials"
	default:
		return category
	}
}

func cleanCities(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}

	cleaned := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		cleaned = append(cleaned, trimmed)
	}
	return cleaned
}
