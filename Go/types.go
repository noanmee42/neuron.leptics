package main

// ClaimsData - структура JSON файла с утверждениями
type ClaimsData struct {
	Timestamp string   `json:"timestamp"`
	Query     string   `json:"query"`
	Response  string   `json:"response"`
	Claims    []string `json:"claims"`
	Count     int      `json:"count"`
}

// FactCheckResult - результат проверки через Jina Grounding API
type FactCheckResult struct {
	Claim      string  `json:"claim"`
	Found      bool    `json:"found"`
	Result     bool    `json:"result"`
	Factuality float64 `json:"factuality"`
	Reason     string  `json:"reason"`
	ReviewURL  string  `json:"review_url,omitempty"`
	Confidence float64 `json:"confidence"`
}

// AnalysisResult - полный результат анализа
type AnalysisResult struct {
	Query            string            `json:"query"`
	Response         string            `json:"response"`
	Claims           []string          `json:"claims"`
	FactCheckResults []FactCheckResult `json:"factcheck_results"`
	Summary          ResultSummary     `json:"summary"`
}

// ResultSummary - сводка результатов
type ResultSummary struct {
	TotalClaims             int `json:"total_claims"`
	ClaimsFound             int `json:"claims_found"`
	ClaimsNotFound          int `json:"claims_not_found"`
	PotentialHallucinations int `json:"potential_hallucinations"`
}
