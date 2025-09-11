package handlers

import (
	"actuworry/backend/models"
	"actuworry/backend/services"
	"encoding/json"
	"net/http"
)

// ActuarialHandler handles HTTP requests for insurance calculations
type ActuarialHandler struct {
	service *services.ActuarialService
}

// NewActuarialHandler creates a handler for actuarial endpoints
func NewActuarialHandler(service *services.ActuarialService) *ActuarialHandler {
	return &ActuarialHandler{
		service: service,
	}
}

// CalculatePremium calculates insurance premium for a single policy
func (h *ActuarialHandler) CalculatePremium(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	// Parse the policy from request body
	var policy models.Policy
	if !parseJSON(w, r, &policy) {
		return
	}

	// Calculate the premium
	result, err := h.service.CalculatePremium(&policy)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Send success response
	sendJSON(w, result, http.StatusOK)
}

// CalculateBatch calculates premiums for multiple policies at once
func (h *ActuarialHandler) CalculateBatch(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var request models.BatchCalculationRequest
	if !parseJSON(w, r, &request) {
		return
	}

	result, err := h.service.CalculateBatch(request.Policies)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	sendJSON(w, result, http.StatusOK)
}

// SensitivityAnalysis shows how premium changes with different inputs
func (h *ActuarialHandler) SensitivityAnalysis(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var request models.SensitivityAnalysisRequest
	if !parseJSON(w, r, &request) {
		return
	}

	result, err := h.service.SensitivityAnalysis(request)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	sendJSON(w, result, http.StatusOK)
}

// PortfolioAnalysis analyzes a group of policies together
func (h *ActuarialHandler) PortfolioAnalysis(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var request models.PortfolioAnalysisRequest
	if !parseJSON(w, r, &request) {
		return
	}

	result, err := h.service.PortfolioAnalysis(request.Policies)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	sendJSON(w, result, http.StatusOK)
}

// GetTables lists available mortality tables (male, female, etc.)
func (h *ActuarialHandler) GetTables(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	tables := h.service.GetAvailableTables()
	sendJSON(w, map[string]interface{}{
		"tables": tables,
		"count":  len(tables),
	}, http.StatusOK)
}

// HealthCheck verifies the service is working properly
func (h *ActuarialHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	tables := h.service.GetAvailableTables()
	sendJSON(w, map[string]interface{}{
		"status":        "healthy",
		"service":       "actuarial",
		"tables_loaded": len(tables),
		"tables":        tables,
	}, http.StatusOK)
}

// ============ Helper Functions ============
// These reduce code duplication and make handlers cleaner

// requireMethod checks if the HTTP method matches what we expect
func requireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		sendError(w, "Method not allowed. Expected "+method, http.StatusMethodNotAllowed)
		return false
	}
	return true
}

// parseJSON reads and validates JSON from request body
func parseJSON(w http.ResponseWriter, r *http.Request, target interface{}) bool {
	err := json.NewDecoder(r.Body).Decode(target)
	if err != nil {
		sendError(w, "Invalid JSON in request body", http.StatusBadRequest)
		return false
	}
	return true
}

// sendJSON sends a successful JSON response
func sendJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// sendError sends an error response in JSON format
func sendError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(models.ErrorResponse{
		Error: message,
	})
}
