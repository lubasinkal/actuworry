package models

// Policy represents a life insurance policy
type Policy struct {
	Age            int     `json:"age" validate:"min=0,max=120"`
	Term           int     `json:"term" validate:"min=0"`
	CoverageAmount float64 `json:"sum_assured" validate:"min=0"`
	InterestRate   float64 `json:"interest_rate" validate:"min=0,max=1"`
	Gender         string  `json:"table_name"`
	ProductType    string  `json:"product_type"`
	SmokerStatus   string  `json:"smoker_status,omitempty"`
	HealthRating   string  `json:"health_rating,omitempty"`
	RatingFactor   float64 `json:"rating_factor,omitempty"`
	DeferralPeriod int     `json:"deferral_period,omitempty"`
}

// PremiumCalculation contains the results of premium calculations
type PremiumCalculation struct {
	NetPremium       float64                `json:"net_premium"`
	GrossPremium     float64                `json:"gross_premium"`
	ReserveSchedule  []float64              `json:"reserve_schedule"`
	ProductType      string                 `json:"product_type"`
	ExpenseDetails   map[string]float64     `json:"expenses,omitempty"`
	AnnualPayout     float64                `json:"annual_payout,omitempty"`
	TotalPremiumCost float64                `json:"total_premium_cost,omitempty"`
	UnderwritingInfo map[string]interface{} `json:"underwriting,omitempty"`
	RiskAssessment   map[string]float64     `json:"risk_assessment,omitempty"`
}

// ExpenseStructure defines expense assumptions for premium calculations
type ExpenseStructure struct {
	InitialExpenseRate float64 `json:"initial_expense_rate"`
	RenewalExpenseRate float64 `json:"renewal_expense_rate"`
	MaintenanceExpense float64 `json:"maintenance_expense"`
	ProfitMargin       float64 `json:"profit_margin"`
}

// BatchCalculationRequest contains multiple policies for batch processing
type BatchCalculationRequest struct {
	Policies []Policy `json:"policies" validate:"required,min=1,max=100"`
}

// BatchCalculationResponse contains results for batch calculations
type BatchCalculationResponse struct {
	Results []PremiumCalculation   `json:"results"`
	Summary map[string]interface{} `json:"summary"`
}

// SensitivityAnalysisRequest defines parameters for sensitivity analysis
type SensitivityAnalysisRequest struct {
	BasePolicy      Policy    `json:"base_policy" validate:"required"`
	InterestRates   []float64 `json:"interest_rates"`
	Ages            []int     `json:"ages,omitempty"`
	CoverageAmounts []float64 `json:"coverage_amounts,omitempty"`
}

// SensitivityResult contains a single sensitivity analysis result
type SensitivityResult struct {
	Parameter string             `json:"parameter"`
	Value     float64            `json:"value"`
	Result    PremiumCalculation `json:"result"`
}

// SensitivityAnalysisResponse contains full sensitivity analysis results
type SensitivityAnalysisResponse struct {
	BaseResult PremiumCalculation        `json:"base_result"`
	Analysis   map[string][]SensitivityResult `json:"analysis"`
}

// PortfolioAnalysisRequest contains policies for portfolio analysis
type PortfolioAnalysisRequest struct {
	Policies []Policy `json:"policies" validate:"required,min=1"`
}

// PortfolioMetrics contains aggregated portfolio statistics
type PortfolioMetrics struct {
	TotalPolicies        int                `json:"total_policies"`
	TotalNetPremium      float64            `json:"total_net_premium"`
	TotalGrossPremium    float64            `json:"total_gross_premium"`
	AverageAge           float64            `json:"average_age"`
	AverageCoverage      float64            `json:"average_coverage"`
	ProductDistribution  map[string]int     `json:"product_distribution"`
	GenderDistribution   map[string]int     `json:"gender_distribution"`
	RiskDistribution     map[string]int     `json:"risk_distribution"`
	ProfitabilityMetrics map[string]float64 `json:"profitability_metrics"`
}

// ErrorResponse standardizes error responses
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}
