// Go/factcheck.go

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// JinaClient - клиент для Jina AI Grounding API
type JinaClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewJinaClient создает новый клиент
func NewJinaClient(apiKey string) *JinaClient {
	return &JinaClient{
		apiKey:  apiKey,
		baseURL: "https://g.jina.ai/",
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// CheckClaim проверяет одно утверждение через Jina Grounding API
func (j *JinaClient) CheckClaim(claim string) (FactCheckResult, error) {
	// Jina Grounding API: GET запрос с утверждением в URL
	// Формат: https://g.jina.ai/YOUR_STATEMENT
	requestURL := j.baseURL + claim

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return FactCheckResult{Claim: claim}, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+j.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "ru-RU, ru")

	resp, err := j.httpClient.Do(req)
	if err != nil {
		return FactCheckResult{Claim: claim}, fmt.Errorf("ошибка запроса: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return FactCheckResult{Claim: claim}, fmt.Errorf("ошибка чтения: %w", err)
	}

	fmt.Printf("   🐛 статус: %d\n", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		return FactCheckResult{Claim: claim}, fmt.Errorf("статус %d: %s", resp.StatusCode, string(body))
	}

	var jinaResponse struct {
		Data struct {
			Factuality float64 `json:"factuality"`
			Result     bool    `json:"result"`
			Reason     string  `json:"reason"`
			References []struct {
				URL          string `json:"url"`
				KeyQuote     string `json:"keyQuote"`
				IsSupportive bool   `json:"isSupportive"`
			} `json:"references"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &jinaResponse); err != nil {
		return FactCheckResult{Claim: claim}, fmt.Errorf("ошибка парсинга: %w", err)
	}

	geminiKey := os.Getenv("GEMINI_API_KEY")
	translatedReason := translateToRussian(jinaResponse.Data.Reason, geminiKey)
	fmt.Printf("   🐛 GEMINI_KEY пустой: %v\n", geminiKey == "")

	var sourceURL, sourceQuote string
	for _, ref := range jinaResponse.Data.References {
		if ref.IsSupportive {
			sourceURL = ref.URL
			sourceQuote = ref.KeyQuote
			break
		}
	}
	if sourceURL == "" && len(jinaResponse.Data.References) > 0 {
		sourceURL = jinaResponse.Data.References[0].URL
		sourceQuote = jinaResponse.Data.References[0].KeyQuote
	}

	return FactCheckResult{
		Claim:      claim,
		Found:      true,
		Result:     jinaResponse.Data.Result,
		Factuality: jinaResponse.Data.Factuality,
		Reason:     translatedReason,
		ReviewURL:  sourceURL,
		KeyQuote:   sourceQuote,
		Confidence: jinaResponse.Data.Factuality,
	}, nil
}

// CheckClaims проверяет список утверждений
func (j *JinaClient) CheckClaims(claims []string) ([]FactCheckResult, error) {
	results := make([]FactCheckResult, 0, len(claims))

	for i, claim := range claims {
		fmt.Printf("   [%d/%d] Проверка: %s\n", i+1, len(claims), claim)

		result, err := j.CheckClaim(claim)
		if err != nil {
			fmt.Printf("   ⚠️  Ошибка: %v\n", err)
			results = append(results, FactCheckResult{
				Claim: claim,
				Found: false,
			})
		} else {
			results = append(results, result)
		}

		if i < len(claims)-1 {
			time.Sleep(500 * time.Millisecond)
		}
	}

	return results, nil
}

// BuildSummary считает сводку по результатам
func BuildSummary(results []FactCheckResult) ResultSummary {
	summary := ResultSummary{
		TotalClaims: len(results),
	}

	for _, r := range results {
		if r.Found && r.Result {
			summary.ClaimsFound++
		} else {
			summary.ClaimsNotFound++
			summary.PotentialHallucinations++
		}
	}

	return summary
}

func translateToRussian(text string, geminiKey string) string {
	if text == "" || geminiKey == "" {
		return text
	}

	requestBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": "Переведи на русский язык, только перевод без пояснений: " + text},
				},
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return text
	}

	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-lite:generateContent?key=" + geminiKey
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("   🐛 GEMINI ошибка запроса: %v\n", err) // <- добавь
		return text
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("   🐛 GEMINI ответ: %s\n", string(body)) // <- добавь

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return text
	}

	if len(result.Candidates) > 0 && len(result.Candidates[0].Content.Parts) > 0 {
		return result.Candidates[0].Content.Parts[0].Text
	}

	return text
}
