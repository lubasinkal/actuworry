package main

import (
	"actuworry/backend/actuarial"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

var mortalityTable actuarial.MortalityTable

func main() {
	var err error
	mortalityTable, err = actuarial.LoadMortalityTable("backend/data/male.csv")
	if err != nil {
		log.Fatalf("failed to load mortality table: %v", err)
	}

	http.HandleFunc("/calculate", calculateHandler)

	fmt.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func calculateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var p actuarial.PolicyHolder
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	netPremium := actuarial.NetPremium(&p, mortalityTable)
	reserves := actuarial.NetPremiumReserves(&p, mortalityTable, netPremium)

	result := actuarial.CalculationResult{
		NetPremium:      netPremium,
		ReserveSchedule: reserves,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}