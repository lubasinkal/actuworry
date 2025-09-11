// Package actuarial provides simple functions for life insurance calculations.
// Think of it as a calculator for insurance premiums and death benefits.
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

// MortalityTable is just a list of death probabilities by age
// Index 0 = probability of death at age 0, Index 50 = probability at age 50, etc.
type MortalityTable []float64

// Policy represents a person's insurance policy details
type Policy struct {
	Age            int     `json:"age"`            // How old is the person?
	Term           int     `json:"term"`           // How many years will the policy last?
	CoverageAmount float64 `json:"sum_assured"`    // How much money paid if person dies?
	InterestRate   float64 `json:"interest_rate"`  // Interest rate for calculations (e.g., 0.05 for 5%)
	Gender         string  `json:"table_name"`     // Male or Female (affects death rates)
	ProductType    string  `json:"product_type"`   // Type of insurance: "term_life" or "whole_life"
	SmokerStatus   string  `json:"smoker_status,omitempty"`   // Does person smoke? Affects risk
	HealthRating   string  `json:"health_rating,omitempty"`   // Health status: "standard", "substandard", "preferred"
	RatingFactor   float64 `json:"rating_factor,omitempty"`   // Risk multiplier (1.0 = normal risk)
	DeferralPeriod int     `json:"deferral_period,omitempty"` // For annuities: years to wait before payments
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

// LoadMortalityTable reads death probability data from a CSV file.
// The CSV should have death rates (qx values) showing probability of death at each age.
// Example: Age 30 might have 0.001 (0.1% chance of death that year)
func LoadMortalityTable(filePath string) (MortalityTable, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open mortality table file: %w", err)
	}
	defer file.Close()

	// Setup CSV reader for tab-delimited files
	csvReader := csv.NewReader(file)
	csvReader.FieldsPerRecord = -1  // Allow variable number of fields
	csvReader.Comma = '\t'           // Tab-delimited

	// Skip the header row
	_, err = csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("could not read CSV header: %w", err)
	}

	// Read all death probabilities
	deathProbabilities := MortalityTable{}
	for {
		row, err := csvReader.Read()
		if err == io.EOF {
			break // End of file reached
		}
		if err != nil {
			return nil, fmt.Errorf("error reading CSV row: %w", err)
		}

		// Death rate is usually in column 3 (index 2)
		if len(row) > 2 {
			deathRateText := strings.TrimSpace(row[2])
			deathRate, err := strconv.ParseFloat(deathRateText, 64)
			
			// If column 3 fails, try column 2 (some formats differ)
			if err != nil {
				deathRateText = strings.TrimSpace(row[1])
				deathRate, err = strconv.ParseFloat(deathRateText, 64)
				if err != nil {
					continue // Skip bad rows
				}
			}
			deathProbabilities = append(deathProbabilities, deathRate)
		}
	}

	return deathProbabilities, nil
}

// CalculatePresentValue tells us what money in the future is worth today.
// Example: $1000 in 5 years at 5% interest is worth less today (about $783)
// Formula: PV = FutureAmount / (1 + interestRate)^years
func CalculatePresentValue(futureAmount float64, interestRate float64, numberOfYears int) float64 {
	// How much the money grows over time
	growthFactor := math.Pow(1+interestRate, float64(numberOfYears))
	
	// Divide to get today's value
	todaysValue := futureAmount / growthFactor
	return todaysValue
}

func CalculateNetPremium(policy *Policy, mortalityTable MortalityTable) float64 {
	if policy.ProductType == "whole_life" {
		return CalculateWholeLifeNetPremium(policy, mortalityTable)
	}
	return CalculateTermLifeNetPremium(policy, mortalityTable)
}

// CalculateTermLifeNetPremium calculates the fair premium for term life insurance.
// It balances what the insurance company expects to pay out vs what they collect.
func CalculateTermLifeNetPremium(policy *Policy, mortalityTable MortalityTable) float64 {
	// Track total expected payouts and premium collections
	expectedPayouts := 0.0
	expectedPremiumsCollected := 0.0

	// Calculate for each year of the policy term
	for yearOfPolicy := 0; yearOfPolicy < policy.Term; yearOfPolicy++ {
		personAge := policy.Age + yearOfPolicy
		
		// Stop if we run out of mortality data
		if personAge >= len(mortalityTable) {
			break
		}

		// Calculate chance person is still alive at start of this year
		chanceStillAlive := calculateSurvivalProbability(policy.Age, yearOfPolicy, mortalityTable)
		
		// Get chance of dying this specific year
		chanceOfDyingThisYear := mortalityTable[personAge]
		
		// Calculate present values (what future money is worth today)
		deathPayoutToday := CalculatePresentValue(policy.CoverageAmount, policy.InterestRate, yearOfPolicy+1)
		premiumToday := CalculatePresentValue(1.0, policy.InterestRate, yearOfPolicy)

		// Add to our running totals
		// Expected payout = chance alive * chance of dying * payout amount
		expectedPayouts += chanceStillAlive * chanceOfDyingThisYear * deathPayoutToday
		
		// Expected premium = chance alive * premium unit
		expectedPremiumsCollected += chanceStillAlive * premiumToday
	}

	// Premium = total expected payouts / total expected premium units
	if expectedPremiumsCollected > 0 {
		return expectedPayouts / expectedPremiumsCollected
	}
	return 0
}

// calculateSurvivalProbability calculates the chance someone survives to a certain year
func calculateSurvivalProbability(startAge int, yearsLater int, mortalityTable MortalityTable) float64 {
	survivalChance := 1.0
	
	// Multiply survival chances for each year
	for year := 0; year < yearsLater; year++ {
		ageThisYear := startAge + year
		chanceOfDying := mortalityTable[ageThisYear]
		chanceOfSurviving := 1.0 - chanceOfDying
		survivalChance *= chanceOfSurviving
	}
	
	return survivalChance
}

// CalculateWholeLifeNetPremium calculates premium for lifetime coverage.
// Unlike term life, this covers until death whenever that happens.
// Person might pay premiums for X years but coverage lasts their whole life.
func CalculateWholeLifeNetPremium(policy *Policy, mortalityTable MortalityTable) float64 {
	expectedPayouts := 0.0
	expectedPremiumsCollected := 0.0

	// Coverage goes until maximum age in our table (usually 100-120 years)
	oldestAgeInTable := len(mortalityTable) - 1
	yearsOfCoverage := oldestAgeInTable - policy.Age
	yearsPayingPremiums := policy.Term // Might pay for 20 years but covered for life

	// Calculate expected costs and premiums year by year
	for yearOfPolicy := 0; yearOfPolicy < yearsOfCoverage; yearOfPolicy++ {
		personAge := policy.Age + yearOfPolicy
		
		if personAge >= len(mortalityTable) {
			break // No more data
		}

		// What's the chance person is still alive this year?
		chanceStillAlive := calculateSurvivalProbability(policy.Age, yearOfPolicy, mortalityTable)
		
		// Death benefit calculation (same as term life)
		chanceOfDyingThisYear := mortalityTable[personAge]
		deathPayoutToday := CalculatePresentValue(policy.CoverageAmount, policy.InterestRate, yearOfPolicy+1)
		expectedPayouts += chanceStillAlive * chanceOfDyingThisYear * deathPayoutToday

		// Premium collection (only during payment period)
		if yearOfPolicy < yearsPayingPremiums {
			premiumToday := CalculatePresentValue(1.0, policy.InterestRate, yearOfPolicy)
			expectedPremiumsCollected += chanceStillAlive * premiumToday
		}
	}

	// Calculate fair premium
	if expectedPremiumsCollected > 0 {
		return expectedPayouts / expectedPremiumsCollected
	}
	return 0
}

// CreateDefaultExpenses returns standard insurance company expense assumptions.
// These cover costs like sales commissions, admin, and profit.
func CreateDefaultExpenses() ExpenseStructure {
	return ExpenseStructure{
		InitialExpenseRate: 0.03,  // 3% of coverage for setting up policy
		RenewalExpenseRate: 0.05,  // 5% of premium for ongoing commission
		MaintenanceExpense: 50.0,   // $50/year for admin costs
		ProfitMargin:       0.15,   // 15% profit margin
	}
}

// CalculateGrossPremium adds company expenses and profit to the net premium.
// Net premium = pure cost of death benefit
// Gross premium = what customer actually pays (includes expenses + profit)
func CalculateGrossPremium(policy *Policy, mortalityTable MortalityTable, netPremium float64, expenses ExpenseStructure) float64 {
	// One-time setup costs spread over policy term
	setupCost := policy.CoverageAmount * expenses.InitialExpenseRate
	setupCostPerYear := setupCost / float64(policy.Term)
	
	// Profit the company wants to make
	profitAmount := netPremium * expenses.ProfitMargin
	
	// Start with net premium plus profit
	grossPremium := netPremium + profitAmount

	// Refine the calculation (iterative because renewal expense depends on premium)
	for i := 0; i < 3; i++ {
		ongoingCommission := grossPremium * expenses.RenewalExpenseRate
		yearlyExpenses := setupCostPerYear + ongoingCommission + expenses.MaintenanceExpense
		grossPremium = netPremium + profitAmount + yearlyExpenses
	}

	// Round to 2 decimal places (cents)
	return math.Round(grossPremium*100) / 100
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

