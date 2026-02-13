// Go/factcheck.go

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// FactCheckAPI - клиент для Google Fact Check API
type FactCheckAPI struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewFactCheckAPI создает новый клиент
func NewFactCheckAPI(apiKey string) *FactCheckAPI {
	return &FactCheckAPI{
		apiKey:  apiKey,
		baseURL: "https://factchecktools.googleapis.com/v1alpha1/claims:search",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// CheckClaim проверяет одно утверждение
func (api *FactCheckAPI) CheckClaim(claim string) (FactCheckResult, error) {
	// Формируем параметры запроса
	params := url.Values{}
	params.Add("query", claim)
	params.Add("key", api.apiKey)
	params.Add("languageCode", "ru")

	// HTTP запрос к Google API
	resp, err := api.httpClient.Get(api.baseURL + "?" + params.Encode())
	if err != nil {
		return FactCheckResult{Claim: claim, Found: false}, fmt.Errorf("ошибка запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return FactCheckResult{Claim: claim, Found: false}, fmt.Errorf("API вернул статус %d: %s", resp.StatusCode, string(body))
	}

	// Парсим ответ Google API
	var apiResponse struct {
		Claims []struct {
			Text        string `json:"text"`
			Claimant    string `json:"claimant"`
			ClaimDate   string `json:"claimDate"`
			ClaimReview []struct {
				Publisher struct {
					Name string `json:"name"`
				} `json:"publisher"`
				URL           string `json:"url"`
				Title         string `json:"title"`
				TextualRating string `json:"textualRating"`
				LanguageCode  string `json:"languageCode"`
			} `json:"claimReview"`
		} `json:"claims"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return FactCheckResult{Claim: claim, Found: false}, fmt.Errorf("ошибка парсинга: %w", err)
	}

	// Если ничего не нашли — возвращаем Found: false
	if len(apiResponse.Claims) == 0 {
		return FactCheckResult{
			Claim:      claim,
			Found:      false,
			Confidence: 0.0,
		}, nil
	}

	// Берём первый результат
	first := apiResponse.Claims[0]

	result := FactCheckResult{
		Claim:        claim,
		Found:        true,
		ClaimantName: first.Claimant,
		Confidence:   0.8,
	}

	// Если есть review — заполняем детали
	if len(first.ClaimReview) > 0 {
		review := first.ClaimReview[0]
		result.TextualRating = review.TextualRating
		result.ReviewPublisher = review.Publisher.Name
		result.ReviewURL = review.URL
		result.ReviewTitle = review.Title
	}

	return result, nil
}

// CheckClaims проверяет список утверждений
func (api *FactCheckAPI) CheckClaims(claims []string) ([]FactCheckResult, error) {
	results := make([]FactCheckResult, 0, len(claims))

	for i, claim := range claims {
		fmt.Printf("   [%d/%d] Проверка: %s\n", i+1, len(claims), claim)

		result, err := api.CheckClaim(claim)
		if err != nil {
			// Не останавливаемся, просто помечаем как не найденное
			fmt.Printf("   ⚠️  Ошибка: %v\n", err)
			results = append(results, FactCheckResult{
				Claim: claim,
				Found: false,
			})
		} else {
			results = append(results, result)
		}

		// Пауза между запросами чтобы не превысить лимиты Google API
		time.Sleep(200 * time.Millisecond)
	}

	return results, nil
}

// BuildSummary считает сводку по результатам проверки
func BuildSummary(results []FactCheckResult) ResultSummary {
	summary := ResultSummary{
		TotalClaims: len(results),
	}

	for _, r := range results {
		if r.Found {
			summary.ClaimsFound++
		} else {
			summary.ClaimsNotFound++
			summary.PotentialHallucinations++
		}
	}

	return summary
}
