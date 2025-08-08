package services

import (
	"actuworry/backend/actuarial"
	"actuworry/backend/models"
	"fmt"
	"strings"
)

// ActuarialService provides actuarial calculation services
type ActuarialService struct {
	mortalityTables map[string]actuarial.MortalityTable
}

// NewActuarialService creates a new actuarial service instance
func NewActuarialService() *ActuarialService {
	return &ActuarialService{
		mortalityTables: make(map[string]actuarial.MortalityTable),
	}
}

// LoadMortalityTable loads a mortality table from file
func (s *ActuarialService) LoadMortalityTable(name, filePath string) error {
	table, err := actuarial.LoadMortalityTable(filePath)
	if err != nil {
		return fmt.Errorf("failed to load mortality table %s: %w", name, err)
	}
	s.mortalityTables[name] = table
	return nil
}

// GetAvailableTables returns list of loaded mortality tables
func (s *ActuarialService) GetAvailableTables() []string {
	tables := make([]string, 0, len(s.mortalityTables))
	for name := range s.mortalityTables {
		tables = append(tables, name)
	}
	return tables
}

// GetMortalityTable retrieves a mortality table by name
func (s *ActuarialService) GetMortalityTable(gender string) (actuarial.MortalityTable, error) {
	tableName := strings.ToLower(gender)
	if tableName == "" {
		tableName = "male"
	}
	
	table, exists := s.mortalityTables[tableName]
	if !exists {
		return nil, fmt.Errorf("mortality table '%s' not found", tableName)
	}
	return table, nil
}

// CalculatePremium calculates premiums for a single policy
func (s *ActuarialService) CalculatePremium(policy *models.Policy) (models.PremiumCalculation, error) {
	// Validate policy
	if err := s.validatePolicy(policy); err != nil {
		return models.PremiumCalculation{}, err
	}
	
	// Get mortality table
	mortalityTable, err := s.GetMortalityTable(policy.Gender)
	if err != nil {
		return models.PremiumCalculation{}, err
	}
	
	// Convert to actuarial.Policy
	actuarialPolicy := s.convertToActuarialPolicy(policy)
	
	// Calculate premium
	result := actuarial.CalculateFullPremium(&actuarialPolicy, mortalityTable)
	
	// Convert back to models.PremiumCalculation
	return s.convertToPremiumCalculation(result), nil
}

// CalculateBatch processes multiple policies
func (s *ActuarialService) CalculateBatch(policies []models.Policy) (models.BatchCalculationResponse, error) {
	if len(policies) == 0 {
		return models.BatchCalculationResponse{}, fmt.Errorf("no policies provided")
	}
	
	if len(policies) > 100 {
		return models.BatchCalculationResponse{}, fmt.Errorf("too many policies (max 100)")
	}
	
	results := make([]models.PremiumCalculation, 0, len(policies))
	totalNetPremium := 0.0
	totalGrossPremium := 0.0
	productCounts := make(map[string]int)
	
	for i, policy := range policies {
		result, err := s.CalculatePremium(&policy)
		if err != nil {
			return models.BatchCalculationResponse{}, 
				fmt.Errorf("failed to calculate policy %d: %w", i+1, err)
		}
		
		results = append(results, result)
		totalNetPremium += result.NetPremium
		totalGrossPremium += result.GrossPremium
		productCounts[result.ProductType]++
	}
	
	summary := map[string]interface{}{
		"total_policies":        len(results),
		"total_net_premium":     totalNetPremium,
		"total_gross_premium":   totalGrossPremium,
		"average_net_premium":   totalNetPremium / float64(len(results)),
		"average_gross_premium": totalGrossPremium / float64(len(results)),
		"product_type_counts":   productCounts,
	}
	
	return models.BatchCalculationResponse{
		Results: results,
		Summary: summary,
	}, nil
}

// SensitivityAnalysis performs sensitivity analysis on a policy
func (s *ActuarialService) SensitivityAnalysis(req models.SensitivityAnalysisRequest) (models.SensitivityAnalysisResponse, error) {
	// Calculate base result
	baseResult, err := s.CalculatePremium(&req.BasePolicy)
	if err != nil {
		return models.SensitivityAnalysisResponse{}, 
			fmt.Errorf("failed to calculate base policy: %w", err)
	}
	
	analysis := make(map[string][]models.SensitivityResult)
	
	// Interest rate sensitivity
	if len(req.InterestRates) > 0 {
		interestResults := make([]models.SensitivityResult, 0, len(req.InterestRates))
		for _, rate := range req.InterestRates {
			testPolicy := req.BasePolicy
			testPolicy.InterestRate = rate
			result, err := s.CalculatePremium(&testPolicy)
			if err != nil {
				continue
			}
			interestResults = append(interestResults, models.SensitivityResult{
				Parameter: "interest_rate",
				Value:     rate,
				Result:    result,
			})
		}
		analysis["interest_rate"] = interestResults
	}
	
	// Age sensitivity
	if len(req.Ages) > 0 {
		ageResults := make([]models.SensitivityResult, 0, len(req.Ages))
		for _, age := range req.Ages {
			testPolicy := req.BasePolicy
			testPolicy.Age = age
			result, err := s.CalculatePremium(&testPolicy)
			if err != nil {
				continue
			}
			ageResults = append(ageResults, models.SensitivityResult{
				Parameter: "age",
				Value:     float64(age),
				Result:    result,
			})
		}
		analysis["age"] = ageResults
	}
	
	// Coverage amount sensitivity
	if len(req.CoverageAmounts) > 0 {
		coverageResults := make([]models.SensitivityResult, 0, len(req.CoverageAmounts))
		for _, coverage := range req.CoverageAmounts {
			testPolicy := req.BasePolicy
			testPolicy.CoverageAmount = coverage
			result, err := s.CalculatePremium(&testPolicy)
			if err != nil {
				continue
			}
			coverageResults = append(coverageResults, models.SensitivityResult{
				Parameter: "coverage_amount",
				Value:     coverage,
				Result:    result,
			})
		}
		analysis["coverage_amount"] = coverageResults
	}
	
	return models.SensitivityAnalysisResponse{
		BaseResult: baseResult,
		Analysis:   analysis,
	}, nil
}

// PortfolioAnalysis analyzes a portfolio of policies
func (s *ActuarialService) PortfolioAnalysis(policies []models.Policy) (models.PortfolioMetrics, error) {
	if len(policies) == 0 {
		return models.PortfolioMetrics{}, fmt.Errorf("no policies provided")
	}
	
	totalAge := 0
	totalCoverage := 0.0
	totalNetPremium := 0.0
	totalGrossPremium := 0.0
	productDist := make(map[string]int)
	genderDist := make(map[string]int)
	riskDist := make(map[string]int)
	
	validPolicies := 0
	for _, policy := range policies {
		result, err := s.CalculatePremium(&policy)
		if err != nil {
			continue
		}
		
		validPolicies++
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
	
	if validPolicies == 0 {
		return models.PortfolioMetrics{}, fmt.Errorf("no valid policies found")
	}
	
	// Calculate profitability metrics
	totalExpectedPayout := totalCoverage * 0.02
	expectedProfit := totalGrossPremium - totalNetPremium
	profitMargin := expectedProfit / totalGrossPremium
	lossRatio := totalExpectedPayout / totalGrossPremium
	
	profitabilityMetrics := map[string]float64{
		"expected_profit":    expectedProfit,
		"profit_margin":      profitMargin,
		"loss_ratio":         lossRatio,
		"expense_ratio":      (totalGrossPremium - totalNetPremium) / totalGrossPremium,
		"combined_ratio":     lossRatio + ((totalGrossPremium - totalNetPremium) / totalGrossPremium),
		"return_on_premium":  expectedProfit / totalNetPremium,
	}
	
	return models.PortfolioMetrics{
		TotalPolicies:        validPolicies,
		TotalNetPremium:      totalNetPremium,
		TotalGrossPremium:    totalGrossPremium,
		AverageAge:           float64(totalAge) / float64(validPolicies),
		AverageCoverage:      totalCoverage / float64(validPolicies),
		ProductDistribution:  productDist,
		GenderDistribution:   genderDist,
		RiskDistribution:     riskDist,
		ProfitabilityMetrics: profitabilityMetrics,
	}, nil
}

// Helper functions

func (s *ActuarialService) validatePolicy(policy *models.Policy) error {
	if policy.Age < 0 || policy.Age > 120 {
		return fmt.Errorf("age must be between 0 and 120")
	}
	if policy.Term < 0 {
		return fmt.Errorf("term must be positive")
	}
	if policy.CoverageAmount <= 0 {
		return fmt.Errorf("coverage amount must be positive")
	}
	if policy.InterestRate < 0 || policy.InterestRate > 1 {
		return fmt.Errorf("interest rate must be between 0 and 1")
	}
	return nil
}

func (s *ActuarialService) convertToActuarialPolicy(policy *models.Policy) actuarial.Policy {
	return actuarial.Policy{
		Age:            policy.Age,
		Term:           policy.Term,
		CoverageAmount: policy.CoverageAmount,
		InterestRate:   policy.InterestRate,
		Gender:         policy.Gender,
		ProductType:    policy.ProductType,
		SmokerStatus:   policy.SmokerStatus,
		HealthRating:   policy.HealthRating,
		RatingFactor:   policy.RatingFactor,
		DeferralPeriod: policy.DeferralPeriod,
	}
}

func (s *ActuarialService) convertToPremiumCalculation(calc actuarial.PremiumCalculation) models.PremiumCalculation {
	return models.PremiumCalculation{
		NetPremium:       calc.NetPremium,
		GrossPremium:     calc.GrossPremium,
		ReserveSchedule:  calc.ReserveSchedule,
		ProductType:      calc.ProductType,
		ExpenseDetails:   calc.ExpenseDetails,
		AnnualPayout:     calc.AnnualPayout,
		TotalPremiumCost: calc.TotalPremiumCost,
		UnderwritingInfo: calc.UnderwritingInfo,
		RiskAssessment:   calc.RiskAssessment,
	}
}
