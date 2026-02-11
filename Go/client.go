// Go/client.go

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// PythonClient - HTTP клиент для взаимодействия с Python API
type PythonClient struct {
	baseURL    string
	httpClient *http.Client
}

// ExtractSaveResponse - ответ от /extract-and-save
type ExtractSaveResponse struct {
	Success     bool     `json:"success"`
	Filename    string   `json:"filename"`
	ClaimsCount int      `json:"claims_count"`
	Claims      []string `json:"claims"`
}

// NewPythonClient создает новый клиент
func NewPythonClient(baseURL string) *PythonClient {
	return &PythonClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// HealthCheck проверяет доступность Python API
func (c *PythonClient) HealthCheck() error {
	resp, err := c.httpClient.Get(c.baseURL + "/health")
	if err != nil {
		return fmt.Errorf("Python API недоступен: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Python API вернул статус %d", resp.StatusCode)
	}

	return nil
}

// ExtractClaims извлекает утверждения из текста (без сохранения)
func (c *PythonClient) ExtractClaims(text string) ([]string, error) {
	requestBody := map[string]string{
		"text": text,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("ошибка сериализации: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.baseURL+"/extract-claims",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ошибка API (статус %d): %s", resp.StatusCode, string(body))
	}

	var response struct {
		Claims []string `json:"claims"`
		Count  int      `json:"count"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа: %w", err)
	}

	return response.Claims, nil
}

// ExtractAndSave извлекает утверждения и сохраняет в JSON файл
func (c *PythonClient) ExtractAndSave(query, response string) (*ExtractSaveResponse, error) {
	requestBody := map[string]string{
		"text":  response,
		"query": query,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("ошибка сериализации: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.baseURL+"/extract-and-save",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ошибка API (статус %d): %s", resp.StatusCode, string(body))
	}

	var result ExtractSaveResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа: %w", err)
	}

	return &result, nil
}
