// Package actuarial provides functions for life insurance calculations.
package actuarial

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
)

type MortalityTable []float64

type Policy struct {
	Age            int     `json:"age"`
	Term           int     `json:"term"`
	CoverageAmount float64 `json:"sum_assured"`
	InterestRate   float64 `json:"interest_rate"`
	Gender         string  `json:"table_name"`
	ProductType    string  `json:"product_type"` // "term_life", "whole_life", "immediate_annuity", "deferred_annuity"
	SmokerStatus   string  `json:"smoker_status,omitempty"` // "smoker", "non_smoker", "unknown"
	HealthRating   string  `json:"health_rating,omitempty"` // "standard", "substandard", "preferred"
	RatingFactor   float64 `json:"rating_factor,omitempty"` // Mortality multiplier (1.0 = standard, >1.0 = substandard)
	DeferralPeriod int     `json:"deferral_period,omitempty"` // Years until annuity payments start
}

type PremiumCalculation struct {
	NetPremium        float64            `json:"net_premium"`
	GrossPremium      float64            `json:"gross_premium"`
	ReserveSchedule   []float64          `json:"reserve_schedule"`
	ProductType       string             `json:"product_type"`
	ExpenseDetails    map[string]float64 `json:"expenses,omitempty"`
	AnnualPayout      float64            `json:"annual_payout,omitempty"`      // For annuities
	TotalPremiumCost  float64            `json:"total_premium_cost,omitempty"` // For annuities
	UnderwritingInfo  map[string]interface{} `json:"underwriting,omitempty"`
	RiskAssessment    map[string]float64 `json:"risk_assessment,omitempty"`
}

type ExpenseStructure struct {
	InitialExpenseRate float64
	RenewalExpenseRate float64
	MaintenanceExpense float64
	ProfitMargin       float64
}

// LoadMortalityTable reads a mortality table from a CSV file into a MortalityTable slice.
// It expects the CSV to have a header row, be tab-delimited, and have the qx value
// in the third column.
func LoadMortalityTable(filePath string) (MortalityTable, error) {
	file, openError := os.Open(filePath)
	if openError != nil {
		return nil, fmt.Errorf("could not open file: %w", openError)
	}
	defer file.Close()

	csvReader := csv.NewReader(file)
	csvReader.FieldsPerRecord = -1
	csvReader.Comma = '\t'

	_, headerError := csvReader.Read()
	if headerError != nil {
		return nil, fmt.Errorf("could not read header: %w", headerError)
	}

	var mortalityRates MortalityTable
	for {
		row, readError := csvReader.Read()
		if readError == io.EOF {
			break
		}
		if readError != nil {
			return nil, fmt.Errorf("could not read row: %w", readError)
		}

		if len(row) > 2 {
			mortalityRateText := strings.TrimSpace(row[2])
			mortalityRate, parseError := strconv.ParseFloat(mortalityRateText, 64)
			if parseError != nil {
				mortalityRateText = strings.TrimSpace(row[1])
				mortalityRate, parseError = strconv.ParseFloat(mortalityRateText, 64)
				if parseError != nil {
					continue
				}
			}
			mortalityRates = append(mortalityRates, mortalityRate)
		}
	}

	return mortalityRates, nil
}

// PresentValue calculates the present value of a single future payment.
func CalculatePresentValue(futureAmount float64, interestRate float64, numberOfYears int) float64 {
	discountFactor := math.Pow(1+interestRate, float64(numberOfYears))
	return futureAmount / discountFactor
}

func CalculateNetPremium(policy *Policy, mortalityTable MortalityTable) float64 {
	if policy.ProductType == "whole_life" {
		return CalculateWholeLifeNetPremium(policy, mortalityTable)
	}
	return CalculateTermLifeNetPremium(policy, mortalityTable)
}

func CalculateTermLifeNetPremium(policy *Policy, mortalityTable MortalityTable) float64 {
	totalExpectedDeathBenefit := 0.0
	totalExpectedPremiumPayments := 0.0

	for year := 0; year < policy.Term; year++ {
		currentAge := policy.Age + year
		if currentAge >= len(mortalityTable) {
			break
		}

		survivalProbability := 1.0
		for previousYear := 0; previousYear < year; previousYear++ {
			survivalProbability *= (1.0 - mortalityTable[policy.Age+previousYear])
		}

		deathProbability := mortalityTable[currentAge]
		deathBenefitPresentValue := CalculatePresentValue(policy.CoverageAmount, policy.InterestRate, year+1)
		premiumPresentValue := CalculatePresentValue(1.0, policy.InterestRate, year)

		totalExpectedDeathBenefit += survivalProbability * deathProbability * deathBenefitPresentValue
		totalExpectedPremiumPayments += survivalProbability * premiumPresentValue
	}

	if totalExpectedPremiumPayments > 0 {
		return totalExpectedDeathBenefit / totalExpectedPremiumPayments
	}
	return 0
}

func CalculateWholeLifeNetPremium(policy *Policy, mortalityTable MortalityTable) float64 {
	totalExpectedDeathBenefit := 0.0
	totalExpectedPremiumPayments := 0.0

	// For whole life, calculate until end of mortality table (lifetime coverage)
	maxAge := len(mortalityTable) - 1
	premiumPayingPeriod := policy.Term // Number of years to pay premiums

	for year := 0; year < maxAge-policy.Age; year++ {
		currentAge := policy.Age + year
		if currentAge >= len(mortalityTable) {
			break
		}

		survivalProbability := 1.0
		for previousYear := 0; previousYear < year; previousYear++ {
			survivalProbability *= (1.0 - mortalityTable[policy.Age+previousYear])
		}

		deathProbability := mortalityTable[currentAge]
		deathBenefitPresentValue := CalculatePresentValue(policy.CoverageAmount, policy.InterestRate, year+1)
		totalExpectedDeathBenefit += survivalProbability * deathProbability * deathBenefitPresentValue

		// Premium payments only during premium paying period
		if year < premiumPayingPeriod {
			premiumPresentValue := CalculatePresentValue(1.0, policy.InterestRate, year)
			totalExpectedPremiumPayments += survivalProbability * premiumPresentValue
		}
	}

	if totalExpectedPremiumPayments > 0 {
		return totalExpectedDeathBenefit / totalExpectedPremiumPayments
	}
	return 0
}

func CreateDefaultExpenses() ExpenseStructure {
	return ExpenseStructure{
		InitialExpenseRate: 0.03,
		RenewalExpenseRate: 0.05,
		MaintenanceExpense: 50.0,
		ProfitMargin:       0.15,
	}
}

func CalculateGrossPremium(policy *Policy, mortalityTable MortalityTable, netPremium float64, expenses ExpenseStructure) float64 {
	initialExpenseAmount := policy.CoverageAmount * expenses.InitialExpenseRate
	profitLoading := netPremium * expenses.ProfitMargin
	basePremium := netPremium + profitLoading

	for iteration := 0; iteration < 3; iteration++ {
		renewalExpenseAmount := basePremium * expenses.RenewalExpenseRate
		totalExpensePerYear := (initialExpenseAmount + renewalExpenseAmount + expenses.MaintenanceExpense) / float64(policy.Term)
		basePremium = netPremium + profitLoading + totalExpensePerYear
	}

	return math.Round(basePremium*100) / 100
}

func CalculateReserveSchedule(policy *Policy, mortalityTable MortalityTable, netPremium float64) []float64 {
	if policy.ProductType == "whole_life" {
		return CalculateWholeLifeReserveSchedule(policy, mortalityTable, netPremium)
	}
	return CalculateTermLifeReserveSchedule(policy, mortalityTable, netPremium)
}

func CalculateTermLifeReserveSchedule(policy *Policy, mortalityTable MortalityTable, netPremium float64) []float64 {
	reserveSchedule := make([]float64, policy.Term+1)

	for currentYear := 0; currentYear <= policy.Term; currentYear++ {
		if currentYear == policy.Term {
			reserveSchedule[currentYear] = 0
			continue
		}

		futureBenefitValue := 0.0
		futurePremiumValue := 0.0
		remainingYears := policy.Term - currentYear
		currentAgeAtYear := policy.Age + currentYear

		for futureYear := 0; futureYear < remainingYears; futureYear++ {
			ageAtFutureYear := currentAgeAtYear + futureYear
			if ageAtFutureYear >= len(mortalityTable) {
				break
			}

			survivalProbability := 1.0
			for yearIndex := 0; yearIndex < futureYear; yearIndex++ {
				survivalProbability *= (1.0 - mortalityTable[currentAgeAtYear+yearIndex])
			}

			deathProbability := mortalityTable[ageAtFutureYear]
			benefitPresentValue := CalculatePresentValue(policy.CoverageAmount, policy.InterestRate, futureYear+1)
			premiumPresentValue := CalculatePresentValue(netPremium, policy.InterestRate, futureYear)

			futureBenefitValue += survivalProbability * deathProbability * benefitPresentValue
			futurePremiumValue += survivalProbability * premiumPresentValue
		}

		reserveSchedule[currentYear] = futureBenefitValue - futurePremiumValue
	}

	return reserveSchedule
}

func CalculateWholeLifeReserveSchedule(policy *Policy, mortalityTable MortalityTable, netPremium float64) []float64 {
	maxAge := len(mortalityTable) - 1
	lifetimeYears := maxAge - policy.Age
	reserveSchedule := make([]float64, lifetimeYears+1)

	for currentYear := 0; currentYear <= lifetimeYears; currentYear++ {
		currentAgeAtYear := policy.Age + currentYear
		if currentAgeAtYear >= len(mortalityTable) {
			break
		}

		futureBenefitValue := 0.0
		futurePremiumValue := 0.0
		remainingLifetimeYears := lifetimeYears - currentYear

		for futureYear := 0; futureYear < remainingLifetimeYears; futureYear++ {
			ageAtFutureYear := currentAgeAtYear + futureYear
			if ageAtFutureYear >= len(mortalityTable) {
				break
			}

			survivalProbability := 1.0
			for yearIndex := 0; yearIndex < futureYear; yearIndex++ {
				survivalProbability *= (1.0 - mortalityTable[currentAgeAtYear+yearIndex])
			}

			deathProbability := mortalityTable[ageAtFutureYear]
			benefitPresentValue := CalculatePresentValue(policy.CoverageAmount, policy.InterestRate, futureYear+1)
			futureBenefitValue += survivalProbability * deathProbability * benefitPresentValue

			// Premium payments only during premium paying period
			if currentYear+futureYear < policy.Term {
				premiumPresentValue := CalculatePresentValue(netPremium, policy.InterestRate, futureYear)
				futurePremiumValue += survivalProbability * premiumPresentValue
			}
		}

		reserveSchedule[currentYear] = futureBenefitValue - futurePremiumValue
	}

	return reserveSchedule
}

// Apply underwriting factors to mortality table
func ApplyUnderwritingFactors(policy *Policy, baseMortalityTable MortalityTable) MortalityTable {
	adjustedTable := make(MortalityTable, len(baseMortalityTable))
	copy(adjustedTable, baseMortalityTable)

	// Apply rating factor
	ratingMultiplier := 1.0
	if policy.RatingFactor > 0 {
		ratingMultiplier = policy.RatingFactor
	} else {
		// Apply standard underwriting factors
		switch policy.SmokerStatus {
		case "smoker":
			ratingMultiplier = 2.0 // Smokers have roughly 2x mortality
		case "non_smoker":
			ratingMultiplier = 0.8 // Non-smokers get a discount
		default:
			ratingMultiplier = 1.0
		}

		switch policy.HealthRating {
		case "preferred":
			ratingMultiplier *= 0.75 // 25% discount for preferred risks
		case "substandard":
			ratingMultiplier *= 1.5 // 50% loading for substandard risks
		default:
			ratingMultiplier *= 1.0
		}
	}

	// Apply the multiplier to all mortality rates, capping at 1.0
	for i, rate := range adjustedTable {
		adjustedTable[i] = math.Min(rate*ratingMultiplier, 1.0)
	}

	return adjustedTable
}

// Calculate immediate annuity premium
func CalculateImmediateAnnuityPremium(policy *Policy, mortalityTable MortalityTable) float64 {
	totalPresentValue := 0.0
	maxAge := len(mortalityTable) - 1

	for year := 0; year < maxAge-policy.Age; year++ {
		currentAge := policy.Age + year
		if currentAge >= len(mortalityTable) {
			break
		}

		survivalProbability := 1.0
		for previousYear := 0; previousYear < year; previousYear++ {
			survivalProbability *= (1.0 - mortalityTable[policy.Age+previousYear])
		}

		annuityPaymentPV := CalculatePresentValue(policy.CoverageAmount, policy.InterestRate, year)
		totalPresentValue += survivalProbability * annuityPaymentPV
	}

	return totalPresentValue
}

// Calculate deferred annuity premium
func CalculateDeferredAnnuityPremium(policy *Policy, mortalityTable MortalityTable) float64 {
	totalPresentValue := 0.0
	maxAge := len(mortalityTable) - 1
	deferralPeriod := policy.DeferralPeriod

	// Calculate survival probability to deferral period
	survivalToDeferral := 1.0
	for year := 0; year < deferralPeriod; year++ {
		currentAge := policy.Age + year
		if currentAge >= len(mortalityTable) {
			return 0
		}
		survivalToDeferral *= (1.0 - mortalityTable[currentAge])
	}

	// Calculate annuity payments starting after deferral period
	for year := deferralPeriod; year < maxAge-policy.Age; year++ {
		currentAge := policy.Age + year
		if currentAge >= len(mortalityTable) {
			break
		}

		survivalProbability := survivalToDeferral
		for previousYear := deferralPeriod; previousYear < year; previousYear++ {
			survivalProbability *= (1.0 - mortalityTable[policy.Age+previousYear])
		}

		annuityPaymentPV := CalculatePresentValue(policy.CoverageAmount, policy.InterestRate, year)
		totalPresentValue += survivalProbability * annuityPaymentPV
	}

	return totalPresentValue
}

// Risk assessment for underwriting
func AssessRisk(policy *Policy, mortalityTable MortalityTable) map[string]float64 {
	baseRate := mortalityTable[policy.Age]
	adjustedTable := ApplyUnderwritingFactors(policy, mortalityTable)
	adjustedRate := adjustedTable[policy.Age]

	return map[string]float64{
		"base_mortality_rate":     baseRate,
		"adjusted_mortality_rate": adjustedRate,
		"risk_multiplier":         adjustedRate / baseRate,
		"annual_death_probability": adjustedRate,
		"expected_lifetime_years":  1.0 / adjustedRate,
	}
}

func CalculateFullPremium(policy *Policy, mortalityTable MortalityTable) PremiumCalculation {
	// Set default product type if not specified
	if policy.ProductType == "" {
		policy.ProductType = "term_life"
	}

	// Apply underwriting factors
	adjustedMortalityTable := ApplyUnderwritingFactors(policy, mortalityTable)
	riskAssessment := AssessRisk(policy, mortalityTable)

	var result PremiumCalculation
	result.ProductType = policy.ProductType
	result.RiskAssessment = riskAssessment

	// Build underwriting info
	underwritingInfo := make(map[string]interface{})
	if policy.SmokerStatus != "" {
		underwritingInfo["smoker_status"] = policy.SmokerStatus
	}
	if policy.HealthRating != "" {
		underwritingInfo["health_rating"] = policy.HealthRating
	}
	if policy.RatingFactor > 0 {
		underwritingInfo["custom_rating_factor"] = policy.RatingFactor
	}
	if len(underwritingInfo) > 0 {
		result.UnderwritingInfo = underwritingInfo
	}

	// Handle different product types
	switch policy.ProductType {
	case "immediate_annuity":
		premiumCost := CalculateImmediateAnnuityPremium(policy, adjustedMortalityTable)
		result.TotalPremiumCost = premiumCost
		result.AnnualPayout = policy.CoverageAmount
		result.NetPremium = premiumCost
		result.GrossPremium = premiumCost * 1.1 // Simple 10% loading for annuities
		return result

	case "deferred_annuity":
		premiumCost := CalculateDeferredAnnuityPremium(policy, adjustedMortalityTable)
		result.TotalPremiumCost = premiumCost
		result.AnnualPayout = policy.CoverageAmount
		result.NetPremium = premiumCost
		result.GrossPremium = premiumCost * 1.1 // Simple 10% loading for annuities
		return result

	default:
		// Life insurance calculations
		netPremium := CalculateNetPremium(policy, adjustedMortalityTable)
		expenseAssumptions := CreateDefaultExpenses()
		grossPremium := CalculateGrossPremium(policy, adjustedMortalityTable, netPremium, expenseAssumptions)
		reserveSchedule := CalculateReserveSchedule(policy, adjustedMortalityTable, netPremium)

		expenseBreakdown := map[string]float64{
			"initial_expense_rate": expenseAssumptions.InitialExpenseRate,
			"renewal_expense_rate": expenseAssumptions.RenewalExpenseRate,
			"maintenance_expense":  expenseAssumptions.MaintenanceExpense,
			"profit_margin":        expenseAssumptions.ProfitMargin,
		}

		result.NetPremium = netPremium
		result.GrossPremium = grossPremium
		result.ReserveSchedule = reserveSchedule
		result.ExpenseDetails = expenseBreakdown
		return result
	}
}

