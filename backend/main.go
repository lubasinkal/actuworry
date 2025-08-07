package main

import (
	"actuworry/backend/actuarial"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

var loadedMortalityTables map[string]actuarial.MortalityTable

type ErrorMessage struct {
	Error string `json:"error"`
}

func sendErrorResponse(responseWriter http.ResponseWriter, errorMessage string, statusCode int) {
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(statusCode)
	json.NewEncoder(responseWriter).Encode(ErrorMessage{Error: errorMessage})
}

func allowCrossOrigin(nextHandler http.HandlerFunc) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, request *http.Request) {
		responseWriter.Header().Set("Access-Control-Allow-Origin", "*")
		responseWriter.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		responseWriter.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if request.Method == "OPTIONS" {
			responseWriter.WriteHeader(http.StatusOK)
			return
		}

		nextHandler(responseWriter, request)
	}
}

func getAvailableTablesHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		sendErrorResponse(responseWriter, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	availableTables := make([]string, 0, len(loadedMortalityTables))
	for tableName := range loadedMortalityTables {
		availableTables = append(availableTables, tableName)
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	json.NewEncoder(responseWriter).Encode(map[string]interface{}{
		"tables": availableTables,
	})
}

func healthCheckHandler(responseWriter http.ResponseWriter, request *http.Request) {
	responseWriter.Header().Set("Content-Type", "application/json")
	json.NewEncoder(responseWriter).Encode(map[string]interface{}{
		"status":        "healthy",
		"tables_loaded": len(loadedMortalityTables),
	})
}

type BatchCalculationRequest struct {
	Policies []actuarial.Policy `json:"policies"`
}

type BatchCalculationResponse struct {
	Results []actuarial.PremiumCalculation `json:"results"`
	Summary map[string]interface{}         `json:"summary"`
}

type SensitivityAnalysisRequest struct {
	BasePolicy      actuarial.Policy `json:"base_policy"`
	InterestRates   []float64        `json:"interest_rates"`
	Ages            []int            `json:"ages,omitempty"`
	CoverageAmounts []float64        `json:"coverage_amounts,omitempty"`
}

type SensitivityResult struct {
	Parameter string  `json:"parameter"`
	Value     float64 `json:"value"`
	Result    actuarial.PremiumCalculation `json:"result"`
}

type SensitivityAnalysisResponse struct {
	BaseResult actuarial.PremiumCalculation `json:"base_result"`
	Analysis   map[string][]SensitivityResult `json:"analysis"`
}

type PortfolioAnalysisRequest struct {
	Policies []actuarial.Policy `json:"policies"`
}

type PortfolioMetrics struct {
	TotalPolicies        int                    `json:"total_policies"`
	TotalNetPremium      float64                `json:"total_net_premium"`
	TotalGrossPremium    float64                `json:"total_gross_premium"`
	AverageAge           float64                `json:"average_age"`
	AverageCoverage      float64                `json:"average_coverage"`
	ProductDistribution  map[string]int         `json:"product_distribution"`
	GenderDistribution   map[string]int         `json:"gender_distribution"`
	RiskDistribution     map[string]int         `json:"risk_distribution"`
	ProfitabilityMetrics map[string]float64     `json:"profitability_metrics"`
}

func calculateBatchHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		sendErrorResponse(responseWriter, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var batchRequest BatchCalculationRequest
	if decodeError := json.NewDecoder(request.Body).Decode(&batchRequest); decodeError != nil {
		sendErrorResponse(responseWriter, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(batchRequest.Policies) == 0 {
		sendErrorResponse(responseWriter, "No policies provided for batch calculation", http.StatusBadRequest)
		return
	}

	if len(batchRequest.Policies) > 100 {
		sendErrorResponse(responseWriter, "Too many policies (max 100 per batch)", http.StatusBadRequest)
		return
	}

	var results []actuarial.PremiumCalculation
	totalNetPremium := 0.0
	totalGrossPremium := 0.0
	productTypeCounts := make(map[string]int)

	for i, policy := range batchRequest.Policies {
		// Validate each policy
		selectedTableName := strings.ToLower(policy.Gender)
		if selectedTableName == "" {
			selectedTableName = "male"
		}

		mortalityTable, tableExists := loadedMortalityTables[selectedTableName]
		if !tableExists {
			sendErrorResponse(responseWriter, fmt.Sprintf("Invalid table_name '%s' for policy %d", policy.Gender, i+1), http.StatusBadRequest)
			return
		}

		if policy.Age < 0 || policy.Term <= 0 || policy.CoverageAmount <= 0 || policy.InterestRate < 0 {
			sendErrorResponse(responseWriter, fmt.Sprintf("Invalid parameters for policy %d", i+1), http.StatusBadRequest)
			return
		}

		if policy.Age+policy.Term >= len(mortalityTable) {
			sendErrorResponse(responseWriter, fmt.Sprintf("Age + term exceeds mortality table length for policy %d", i+1), http.StatusBadRequest)
			return
		}

		calculationResult := actuarial.CalculateFullPremium(&policy, mortalityTable)
		results = append(results, calculationResult)

		totalNetPremium += calculationResult.NetPremium
		totalGrossPremium += calculationResult.GrossPremium
		productTypeCounts[calculationResult.ProductType]++
	}

	summary := map[string]interface{}{
		"total_policies":        len(results),
		"total_net_premium":     totalNetPremium,
		"total_gross_premium":   totalGrossPremium,
		"average_net_premium":   totalNetPremium / float64(len(results)),
		"average_gross_premium": totalGrossPremium / float64(len(results)),
		"product_type_counts":   productTypeCounts,
	}

	response := BatchCalculationResponse{
		Results: results,
		Summary: summary,
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	if encodeError := json.NewEncoder(responseWriter).Encode(response); encodeError != nil {
		log.Printf("Failed to encode batch response: %v", encodeError)
		sendErrorResponse(responseWriter, "Failed to encode response", http.StatusInternalServerError)
	}
}

func sensitivityAnalysisHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		sendErrorResponse(responseWriter, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var sensRequest SensitivityAnalysisRequest
	if decodeError := json.NewDecoder(request.Body).Decode(&sensRequest); decodeError != nil {
		sendErrorResponse(responseWriter, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get base mortality table
	selectedTableName := strings.ToLower(sensRequest.BasePolicy.Gender)
	if selectedTableName == "" {
		selectedTableName = "male"
	}

	mortalityTable, tableExists := loadedMortalityTables[selectedTableName]
	if !tableExists {
		sendErrorResponse(responseWriter, fmt.Sprintf("Invalid table_name: '%s'", sensRequest.BasePolicy.Gender), http.StatusBadRequest)
		return
	}

	// Calculate base result
	baseResult := actuarial.CalculateFullPremium(&sensRequest.BasePolicy, mortalityTable)

	analysis := make(map[string][]SensitivityResult)

	// Interest rate sensitivity
	if len(sensRequest.InterestRates) > 0 {
		var interestResults []SensitivityResult
		for _, rate := range sensRequest.InterestRates {
			testPolicy := sensRequest.BasePolicy
			testPolicy.InterestRate = rate
			result := actuarial.CalculateFullPremium(&testPolicy, mortalityTable)
			interestResults = append(interestResults, SensitivityResult{
				Parameter: "interest_rate",
				Value:     rate,
				Result:    result,
			})
		}
		analysis["interest_rate"] = interestResults
	}

	// Age sensitivity
	if len(sensRequest.Ages) > 0 {
		var ageResults []SensitivityResult
		for _, age := range sensRequest.Ages {
			testPolicy := sensRequest.BasePolicy
			testPolicy.Age = age
			result := actuarial.CalculateFullPremium(&testPolicy, mortalityTable)
			ageResults = append(ageResults, SensitivityResult{
				Parameter: "age",
				Value:     float64(age),
				Result:    result,
			})
		}
		analysis["age"] = ageResults
	}

	// Coverage amount sensitivity
	if len(sensRequest.CoverageAmounts) > 0 {
		var coverageResults []SensitivityResult
		for _, coverage := range sensRequest.CoverageAmounts {
			testPolicy := sensRequest.BasePolicy
			testPolicy.CoverageAmount = coverage
			result := actuarial.CalculateFullPremium(&testPolicy, mortalityTable)
			coverageResults = append(coverageResults, SensitivityResult{
				Parameter: "coverage_amount",
				Value:     coverage,
				Result:    result,
			})
		}
		analysis["coverage_amount"] = coverageResults
	}

	response := SensitivityAnalysisResponse{
		BaseResult: baseResult,
		Analysis:   analysis,
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	if encodeError := json.NewEncoder(responseWriter).Encode(response); encodeError != nil {
		log.Printf("Failed to encode sensitivity response: %v", encodeError)
		sendErrorResponse(responseWriter, "Failed to encode response", http.StatusInternalServerError)
	}
}

func portfolioAnalysisHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		sendErrorResponse(responseWriter, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var portfolioRequest PortfolioAnalysisRequest
	if decodeError := json.NewDecoder(request.Body).Decode(&portfolioRequest); decodeError != nil {
		sendErrorResponse(responseWriter, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(portfolioRequest.Policies) == 0 {
		sendErrorResponse(responseWriter, "No policies provided for portfolio analysis", http.StatusBadRequest)
		return
	}

	// Calculate all policies
	var results []actuarial.PremiumCalculation
	totalAge := 0
	totalCoverage := 0.0
	totalNetPremium := 0.0
	totalGrossPremium := 0.0
	productDist := make(map[string]int)
	genderDist := make(map[string]int)
	riskDist := make(map[string]int)

	for _, policy := range portfolioRequest.Policies {
		selectedTableName := strings.ToLower(policy.Gender)
		if selectedTableName == "" {
			selectedTableName = "male"
		}

		mortalityTable, tableExists := loadedMortalityTables[selectedTableName]
		if !tableExists {
			continue // Skip invalid policies
		}

		result := actuarial.CalculateFullPremium(&policy, mortalityTable)
		results = append(results, result)

		// Accumulate metrics
		totalAge += policy.Age
		totalCoverage += policy.CoverageAmount
		totalNetPremium += result.NetPremium
		totalGrossPremium += result.GrossPremium
		productDist[result.ProductType]++
		genderDist[policy.Gender]++

		// Risk categorization
		if policy.SmokerStatus == "smoker" || policy.HealthRating == "substandard" {
			riskDist["high_risk"]++
		} else if policy.HealthRating == "preferred" || policy.SmokerStatus == "non_smoker" {
			riskDist["low_risk"]++
		} else {
			riskDist["standard_risk"]++
		}
	}

	policyCount := len(results)
	if policyCount == 0 {
		sendErrorResponse(responseWriter, "No valid policies found", http.StatusBadRequest)
		return
	}

	// Calculate profitability metrics
	totalExpectedPayout := totalCoverage * 0.02 // Assume 2% mortality rate for simplification
	expectedProfit := totalGrossPremium - totalNetPremium
	profitMargin := expectedProfit / totalGrossPremium
	lossRatio := totalExpectedPayout / totalGrossPremium

	profitabilityMetrics := map[string]float64{
		"expected_profit":       expectedProfit,
		"profit_margin":         profitMargin,
		"loss_ratio":            lossRatio,
		"expense_ratio":         (totalGrossPremium - totalNetPremium) / totalGrossPremium,
		"combined_ratio":        lossRatio + ((totalGrossPremium - totalNetPremium) / totalGrossPremium),
		"return_on_premium":     expectedProfit / totalNetPremium,
	}

	metrics := PortfolioMetrics{
		TotalPolicies:        policyCount,
		TotalNetPremium:      totalNetPremium,
		TotalGrossPremium:    totalGrossPremium,
		AverageAge:           float64(totalAge) / float64(policyCount),
		AverageCoverage:      totalCoverage / float64(policyCount),
		ProductDistribution:  productDist,
		GenderDistribution:   genderDist,
		RiskDistribution:     riskDist,
		ProfitabilityMetrics: profitabilityMetrics,
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	if encodeError := json.NewEncoder(responseWriter).Encode(metrics); encodeError != nil {
		log.Printf("Failed to encode portfolio response: %v", encodeError)
		sendErrorResponse(responseWriter, "Failed to encode response", http.StatusInternalServerError)
	}
}

func calculatePremiumHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		sendErrorResponse(responseWriter, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var policyRequest actuarial.Policy
	if decodeError := json.NewDecoder(request.Body).Decode(&policyRequest); decodeError != nil {
		sendErrorResponse(responseWriter, "Invalid request body", http.StatusBadRequest)
		return
	}

	selectedTableName := strings.ToLower(policyRequest.Gender)
	if selectedTableName == "" {
		selectedTableName = "male"
	}

	mortalityTable, tableExists := loadedMortalityTables[selectedTableName]
	if !tableExists {
		sendErrorResponse(responseWriter, fmt.Sprintf("Invalid table_name: '%s'", policyRequest.Gender), http.StatusBadRequest)
		return
	}

	if policyRequest.Age < 0 {
		sendErrorResponse(responseWriter, "Age cannot be negative", http.StatusBadRequest)
		return
	}
	if policyRequest.Term <= 0 {
		sendErrorResponse(responseWriter, "Term must be positive", http.StatusBadRequest)
		return
	}
	if policyRequest.CoverageAmount <= 0 {
		sendErrorResponse(responseWriter, "Coverage amount must be positive", http.StatusBadRequest)
		return
	}
	if policyRequest.InterestRate < 0 {
		sendErrorResponse(responseWriter, "Interest rate cannot be negative", http.StatusBadRequest)
		return
	}
	if policyRequest.Age+policyRequest.Term >= len(mortalityTable) {
		sendErrorResponse(responseWriter, "Age + term exceeds mortality table length", http.StatusBadRequest)
		return
	}

	calculationResult := actuarial.CalculateFullPremium(&policyRequest, mortalityTable)

	responseWriter.Header().Set("Content-Type", "application/json")
	if encodeError := json.NewEncoder(responseWriter).Encode(calculationResult); encodeError != nil {
		log.Printf("Failed to encode response: %v", encodeError)
		sendErrorResponse(responseWriter, "Failed to encode response", http.StatusInternalServerError)
	}
}

func main() {
	loadedMortalityTables = make(map[string]actuarial.MortalityTable)

	tableNames := []string{"male", "female"}
	for _, tableName := range tableNames {
		filePath := fmt.Sprintf("backend/data/%s.csv", tableName)
		mortalityTable, loadError := actuarial.LoadMortalityTable(filePath)
		if loadError != nil {
			log.Fatalf("failed to load mortality table '%s': %v", tableName, loadError)
		}
		loadedMortalityTables[tableName] = mortalityTable
		log.Printf("Successfully loaded mortality table: %s", tableName)
	}

	http.HandleFunc("/calculate", allowCrossOrigin(calculatePremiumHandler))
	http.HandleFunc("/calculate/batch", allowCrossOrigin(calculateBatchHandler))
	http.HandleFunc("/calculate/sensitivity", allowCrossOrigin(sensitivityAnalysisHandler))
	http.HandleFunc("/analyze/portfolio", allowCrossOrigin(portfolioAnalysisHandler))
	http.HandleFunc("/tables", allowCrossOrigin(getAvailableTablesHandler))
	http.HandleFunc("/health", allowCrossOrigin(healthCheckHandler))

	staticFileServer := http.FileServer(http.Dir("frontend/"))
	http.Handle("/", http.StripPrefix("/", staticFileServer))

	fmt.Println("Actuworry Server starting on port 8080...")
	fmt.Println("API available at: http://localhost:8080/calculate")
	fmt.Println("Frontend available at: http://localhost:8080")

	if serverError := http.ListenAndServe(":8080", nil); serverError != nil {
		log.Fatalf("failed to start server: %v", serverError)
	}
}
