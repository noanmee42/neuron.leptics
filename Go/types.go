// Go/types.go

package main


// ClaimsData - структура JSON файла с утверждениями
type ClaimsData struct {
	Timestamp string   `json:"timestamp"`
	Query     string   `json:"query"`
	Response  string   `json:"response"`
	Claims    []string `json:"claims"`
	Count     int      `json:"count"`
}

// FactCheckResult - результат проверки через Google Fact Check API
type FactCheckResult struct {
	Claim           string  `json:"claim"`
	Found           bool    `json:"found"`
	TextualRating   string  `json:"textual_rating,omitempty"`
	ReviewPublisher string  `json:"review_publisher,omitempty"`
	ReviewURL       string  `json:"review_url,omitempty"`
	ReviewTitle     string  `json:"review_title,omitempty"`
	ClaimantName    string  `json:"claimant_name,omitempty"`
	Confidence      float64 `json:"confidence"`
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

// VerificationResult - старая структура (пока не удаляем, может пригодиться)
type VerificationResult struct {
	Claim           string  `json:"claim"`
	IsHallucination bool    `json:"is_hallucination"`
	Confidence      float64 `json:"confidence"`
	Explanation     string  `json:"explanation"`
	Source          string  `json:"source"`
}
