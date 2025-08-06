package main

import (
	"actuworry/backend/actuarial"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// mortalityTables holds the pre-loaded mortality data.
var mortalityTables map[string]actuarial.MortalityTable

// ErrorResponse is a structured error message for the API.
type ErrorResponse struct {
	Error string `json:"error"`
}

// writeError sends a JSON error response.
func writeError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func main() {
	// Pre-load all mortality tables into memory.
	mortalityTables = make(map[string]actuarial.MortalityTable)
	
	tablesToLoad := []string{"male", "female"} // Add more table names here
	for _, name := range tablesToLoad {
		path := fmt.Sprintf("backend/data/%s.csv", name)
		table, err := actuarial.LoadMortalityTable(path)
		if err != nil {
			log.Fatalf("failed to load mortality table '%s': %v", name, err)
		}
		mortalityTables[name] = table
		log.Printf("Successfully loaded mortality table: %s", name)
	}

	http.HandleFunc("/calculate", calculateHandler)

	fmt.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func calculateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var p actuarial.PolicyHolder
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// --- Input Validation ---
	tableName := strings.ToLower(p.TableName)
	if tableName == "" {
		tableName = "male" // Default to male table if not provided
	}

	table, ok := mortalityTables[tableName]
	if !ok {
		writeError(w, fmt.Sprintf("Invalid table_name: '%s'", p.TableName), http.StatusBadRequest)
		return
	}

	if p.Age < 0 {
		writeError(w, "Invalid input: age cannot be negative", http.StatusBadRequest)
		return
	}
	if p.Term <= 0 {
		writeError(w, "Invalid input: term must be positive", http.StatusBadRequest)
		return
	}
	if p.SumAssured <= 0 {
		writeError(w, "Invalid input: sum_assured must be positive", http.StatusBadRequest)
		return
	}
	if p.InterestRate < 0 {
		writeError(w, "Invalid input: interest_rate cannot be negative", http.StatusBadRequest)
		return
	}
	if p.Age+p.Term >= len(table) {
		writeError(w, "Invalid input: age + term exceeds mortality table length", http.StatusBadRequest)
		return
	}
	// --- End Validation ---

	netPremium := actuarial.NetPremium(&p, table)
	reserves := actuarial.NetPremiumReserves(&p, table, netPremium)

	result := actuarial.CalculationResult{
		NetPremium:      netPremium,
		ReserveSchedule: reserves,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Failed to encode response: %v", err)
		writeError(w, "Failed to encode response", http.StatusInternalServerError)
	}
}