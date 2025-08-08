package handlers

import (
	"actuworry/backend/models"
	"actuworry/backend/services"
	"encoding/json"
	"net/http"
)

// ActuarialHandler handles actuarial-related HTTP requests
type ActuarialHandler struct {
	service *services.ActuarialService
}

// NewActuarialHandler creates a new actuarial handler
func NewActuarialHandler(service *services.ActuarialService) *ActuarialHandler {
	return &ActuarialHandler{
		service: service,
	}
}

// CalculatePremium handles single premium calculation requests
func (h *ActuarialHandler) CalculatePremium(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var policy models.Policy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	result, err := h.service.CalculatePremium(&policy)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	sendJSON(w, result, http.StatusOK)
}

// CalculateBatch handles batch premium calculation requests
func (h *ActuarialHandler) CalculateBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var request models.BatchCalculationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	result, err := h.service.CalculateBatch(request.Policies)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	sendJSON(w, result, http.StatusOK)
}

// SensitivityAnalysis handles sensitivity analysis requests
func (h *ActuarialHandler) SensitivityAnalysis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var request models.SensitivityAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	result, err := h.service.SensitivityAnalysis(request)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	sendJSON(w, result, http.StatusOK)
}

// PortfolioAnalysis handles portfolio analysis requests
func (h *ActuarialHandler) PortfolioAnalysis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var request models.PortfolioAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	result, err := h.service.PortfolioAnalysis(request.Policies)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	sendJSON(w, result, http.StatusOK)
}

// GetTables returns available mortality tables
func (h *ActuarialHandler) GetTables(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	tables := h.service.GetAvailableTables()
	sendJSON(w, map[string]interface{}{
		"tables": tables,
		"count":  len(tables),
	}, http.StatusOK)
}

// HealthCheck returns service health status
func (h *ActuarialHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
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

// Helper functions

func sendJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func sendError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(models.ErrorResponse{
		Error: message,
	})
}
