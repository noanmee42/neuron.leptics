// Go/factcheck.go

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

type JinaClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewJinaClient(apiKey string) *JinaClient {
	return &JinaClient{
		apiKey:  apiKey,
		baseURL: "https://g.jina.ai/",
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (j *JinaClient) checkViaPost(claim string) ([]byte, int, error) {
	type jinaRequest struct {
		Statement string `json:"statement"`
	}
	jsonData, err := json.Marshal(jinaRequest{Statement: claim})
	if err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequest("POST", j.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+j.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := j.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	return body, resp.StatusCode, err
}

func (j *JinaClient) checkViaGet(claim string) ([]byte, int, error) {
	requestURL := j.baseURL + url.PathEscape(claim)

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+j.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := j.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	return body, resp.StatusCode, err
}

func (j *JinaClient) CheckClaim(claim string) (FactCheckResult, error) {
	// Сначала пробуем POST с json.Marshal — корректно передаёт кириллицу
	body, status, err := j.checkViaPost(claim)
	if err != nil {
		return FactCheckResult{Claim: claim}, fmt.Errorf("ошибка POST запроса: %w", err)
	}

	// Если POST вернул 422 — fallback на GET (для английских утверждений)
	if status == 422 {
		fmt.Printf("   ⚠️  POST вернул 422, пробуем GET...\n")
		body, status, err = j.checkViaGet(claim)
		if err != nil {
			return FactCheckResult{Claim: claim}, fmt.Errorf("ошибка GET запроса: %w", err)
		}
	}

	if status != http.StatusOK {
		return FactCheckResult{Claim: claim}, fmt.Errorf("статус %d: %s", status, string(body))
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

func (j *JinaClient) CheckClaims(claims []string) ([]FactCheckResult, error) {
	results := make([]FactCheckResult, 0, len(claims))

	for i, claim := range claims {
		fmt.Printf("   [%d/%d] Проверка: %s\n", i+1, len(claims), claim)

		result, err := j.CheckClaim(claim)
		if err != nil {
			fmt.Printf("   ⚠️  Ошибка: %v\n", err)
			results = append(results, FactCheckResult{Claim: claim, Found: false})
		} else {
			results = append(results, result)
		}

		if i < len(claims)-1 {
			time.Sleep(500 * time.Millisecond)
		}
	}

	return results, nil
}

func BuildSummary(results []FactCheckResult) ResultSummary {
	summary := ResultSummary{TotalClaims: len(results)}

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
		return text
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

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
