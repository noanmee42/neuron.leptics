// Go/factcheck.go

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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
	payload := fmt.Sprintf(`{"statement": %q, "lang": "ru"}`, claim)
	req, err := http.NewRequest("POST", j.baseURL, strings.NewReader(payload))
	if err != nil {
		return FactCheckResult{Claim: claim}, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+j.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := j.httpClient.Do(req)
	if err != nil {
		return FactCheckResult{Claim: claim}, fmt.Errorf("ошибка запроса: %w", err)
	}
	defer resp.Body.Close()

	// Читаем тело ОДИН РАЗ сразу — до любых проверок статуса
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return FactCheckResult{Claim: claim}, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	// Временный дебаг — показываем сырой ответ
	fmt.Printf("   🐛 RAW ответ Jina: %s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		return FactCheckResult{Claim: claim}, fmt.Errorf("Jina API вернул статус %d: %s", resp.StatusCode, string(body))
	}

	// Парсинг ответа Jina
	var jinaResponse struct {
		Code   int `json:"code"`
		Status int `json:"status"`
		Data   struct {
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

	// Собираем первую поддерживающую ссылку
	var supportingURL string
	for _, ref := range jinaResponse.Data.References {
		if ref.IsSupportive {
			supportingURL = ref.URL
			break
		}
	}

	return FactCheckResult{
		Claim:      claim,
		Found:      true,
		Result:     jinaResponse.Data.Result,
		Factuality: jinaResponse.Data.Factuality,
		Reason:     jinaResponse.Data.Reason,
		ReviewURL:  supportingURL,
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
