package handlers

import (
	"actuworry/backend/actuarial"
	"actuworry/backend/models"
	"actuworry/backend/services"
	"encoding/json"
	"net/http"
)

type ActuarialHandler struct {
	service *services.ActuarialService
}

func NewActuarialHandler(service *services.ActuarialService) *ActuarialHandler {
	return &ActuarialHandler{service: service}
}

func (h *ActuarialHandler) CalculatePremium(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var policy models.Policy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		sendError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	result, err := h.service.CalculatePremium(&policy)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}
	sendJSON(w, result, http.StatusOK)
}

func (h *ActuarialHandler) CalculateBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var request models.BatchCalculationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	result, err := h.service.CalculateBatch(request.Policies)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}
	sendJSON(w, result, http.StatusOK)
}

func (h *ActuarialHandler) SensitivityAnalysis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var request models.SensitivityAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	result, err := h.service.SensitivityAnalysis(request)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}
	sendJSON(w, result, http.StatusOK)
}

func (h *ActuarialHandler) PortfolioAnalysis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var request models.PortfolioAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	result, err := h.service.PortfolioAnalysis(request.Policies)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}
	sendJSON(w, result, http.StatusOK)
}

func (h *ActuarialHandler) GetTables(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	tables := h.service.GetAvailableTables()
	sendJSON(w, map[string]interface{}{"tables": tables, "count": len(tables)}, http.StatusOK)
}

func (h *ActuarialHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	tables := h.service.GetAvailableTables()
	sendJSON(w, map[string]interface{}{"status": "healthy", "service": "actuarial", "tables_loaded": len(tables), "tables": tables}, http.StatusOK)
}

// v-star Advanced Features

func (h *ActuarialHandler) MonteCarloSimulation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		NumPaths int     `json:"num_paths"`
		Drift    float64 `json:"drift"`
		Vol      float64 `json:"volatility"`
		Seed     uint64  `json:"seed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.NumPaths <= 0 {
		req.NumPaths = 10000
	}
	if req.Drift <= 0 {
		req.Drift = 0.02
	}
	if req.Vol <= 0 {
		req.Vol = 0.15
	}
	result := actuarial.RunMCWithRisk(req.NumPaths, 1000000, req.Drift, req.Vol, req.Seed)
	sendJSON(w, result, http.StatusOK)
}

func (h *ActuarialHandler) RiskAnalysis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Losees []float64 `json:"losses"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	result := actuarial.ComputeRiskReport(req.Losees)
	sendJSON(w, result, http.StatusOK)
}

func (h *ActuarialHandler) DurationCalculator(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		CashFlows []float64 `json:"cash_flows"`
		Rate      float64   `json:"rate"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	mac, mod, conv := actuarial.ComputeDuration(req.CashFlows, req.Rate)
	sendJSON(w, map[string]float64{"macaulay_duration": mac, "modified_duration": mod, "convexity": conv}, http.StatusOK)
}

func (h *ActuarialHandler) RateConverterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Nominal     float64 `json:"nominal_rate"`
		Effective   float64 `json:"effective_rate"`
		Compounding int     `json:"compounding"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Compounding <= 0 {
		req.Compounding = 12
	}
	result := map[string]float64{}
	if req.Nominal > 0 {
		result["nominal_rate"] = req.Nominal
		result["effective_rate"] = actuarial.NominalToEffective(req.Nominal, req.Compounding)
		result["force_of_interest"] = actuarial.ForceOfInterest(req.Nominal)
	} else if req.Effective > 0 {
		result["nominal_rate"] = actuarial.EffectiveToNominal(req.Effective, req.Compounding)
		result["effective_rate"] = req.Effective
		result["force_of_interest"] = actuarial.ForceOfInterest(req.Effective)
	} else {
		sendError(w, "Provide either nominal_rate or effective_rate", http.StatusBadRequest)
		return
	}
	sendJSON(w, result, http.StatusOK)
}

func (h *ActuarialHandler) EndowmentCalculator(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Age        int     `json:"age"`
		Term       int     `json:"term"`
		SumAssured float64 `json:"sum_assured"`
		Rate       float64 `json:"rate"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	mortTable, _ := h.service.GetMortalityTable("male")
	result := actuarial.CalcEndowmentNSP(req.Age, req.Term, req.SumAssured, mortTable, req.Rate)
	sendJSON(w, result, http.StatusOK)
}

func (h *ActuarialHandler) RetrospectiveReserve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Policy models.Policy `json:"policy"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	mortTable, _ := h.service.GetMortalityTable(req.Policy.Gender)
	netPrem, _ := h.service.CalculatePremium(&req.Policy)
	result := actuarial.CalcRetrospectiveReserve(&actuarial.Policy{
		Age:            req.Policy.Age,
		Term:           req.Policy.Term,
		CoverageAmount: req.Policy.CoverageAmount,
		InterestRate:   req.Policy.InterestRate,
	}, mortTable, netPrem.NetPremium)
	sendJSON(w, result, http.StatusOK)
}

func (h *ActuarialHandler) BondValuation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Face   float64 `json:"face_value"`
		Coupon float64 `json:"coupon_rate"`
		Years  int     `json:"years"`
		YTM    float64 `json:"yield_to_maturity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	result := actuarial.ValueBond(req.Face, req.Coupon, req.Years, req.YTM)
	sendJSON(w, result, http.StatusOK)
}

// Helpers
func sendJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func sendError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(models.ErrorResponse{Error: message})
}
