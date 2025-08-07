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
