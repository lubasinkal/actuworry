package actuarial

import (
	"fmt"
	"testing"
)

// TestPresentValue shows how present value calculation works
func TestPresentValue(t *testing.T) {
	// Simple example: What's $1000 in 5 years worth today at 5% interest?
	futureAmount := 1000.0
	interestRate := 0.05 // 5%
	years := 5
	
	presentValue := CalculatePresentValue(futureAmount, interestRate, years)
	
	// Expected: around $783.53
	fmt.Printf("$%.2f in %d years is worth $%.2f today at %.1f%% interest\n", 
		futureAmount, years, presentValue, interestRate*100)
	
	if presentValue < 700 || presentValue > 800 {
		t.Errorf("Present value seems wrong: got %.2f", presentValue)
	}
}

// TestTermLifePremium shows basic term life insurance calculation
func TestTermLifePremium(t *testing.T) {
	// Create a simple mortality table
	// These are fake numbers for demonstration
	mortalityTable := MortalityTable{
		// Age 0-29: very low death rates
		0.001, 0.001, 0.001, 0.001, 0.001, 0.001, 0.001, 0.001, 0.001, 0.001,
		0.001, 0.001, 0.001, 0.001, 0.001, 0.001, 0.001, 0.001, 0.001, 0.001,
		0.001, 0.001, 0.001, 0.001, 0.001, 0.001, 0.001, 0.001, 0.001, 0.001,
		// Age 30-39: slightly higher
		0.002, 0.002, 0.002, 0.002, 0.002, 0.003, 0.003, 0.003, 0.003, 0.004,
		// Age 40-49: increasing
		0.005, 0.005, 0.006, 0.006, 0.007, 0.008, 0.009, 0.010, 0.011, 0.012,
		// Age 50-59: higher still
		0.014, 0.016, 0.018, 0.020, 0.023, 0.026, 0.029, 0.033, 0.037, 0.042,
		// Age 60+: much higher
		0.048, 0.055, 0.063, 0.072, 0.082, 0.094, 0.107, 0.122, 0.139, 0.158,
	}
	
	// Create a policy for a 30-year-old
	policy := Policy{
		Age:            30,
		Term:           20,        // 20-year term life insurance
		CoverageAmount: 100000,    // $100,000 death benefit
		InterestRate:   0.05,      // 5% discount rate
		ProductType:    "term_life",
	}
	
	// Calculate the premium
	premium := CalculateTermLifeNetPremium(&policy, mortalityTable)
	
	fmt.Printf("\n=== Term Life Insurance Example ===\n")
	fmt.Printf("Person age: %d\n", policy.Age)
	fmt.Printf("Coverage term: %d years\n", policy.Term)
	fmt.Printf("Death benefit: $%.0f\n", policy.CoverageAmount)
	fmt.Printf("Calculated annual premium: $%.2f\n", premium)
	fmt.Printf("Monthly premium: $%.2f\n", premium/12)
	
	// Basic sanity check
	if premium < 0 || premium > 10000 {
		t.Errorf("Premium seems unrealistic: %.2f", premium)
	}
}

// TestGrossPremium shows how expenses are added to net premium
func TestGrossPremium(t *testing.T) {
	netPremium := 500.0 // $500 net premium (pure cost)
	
	policy := Policy{
		Age:            35,
		Term:           20,
		CoverageAmount: 100000,
		InterestRate:   0.05,
	}
	
	expenses := CreateDefaultExpenses()
	
	grossPremium := CalculateGrossPremium(&policy, nil, netPremium, expenses)
	
	fmt.Printf("\n=== Premium Loading Example ===\n")
	fmt.Printf("Net premium (pure cost): $%.2f\n", netPremium)
	fmt.Printf("Setup cost (%.1f%% of coverage): $%.2f\n", 
		expenses.InitialExpenseRate*100, 
		policy.CoverageAmount*expenses.InitialExpenseRate)
	fmt.Printf("Profit margin (%.1f%%): $%.2f\n", 
		expenses.ProfitMargin*100, 
		netPremium*expenses.ProfitMargin)
	fmt.Printf("Final gross premium: $%.2f\n", grossPremium)
	fmt.Printf("Total loading: %.1f%%\n", (grossPremium-netPremium)/netPremium*100)
	
	if grossPremium < netPremium {
		t.Errorf("Gross premium should be higher than net premium")
	}
}

// TestSurvivalProbability demonstrates survival calculation
func TestSurvivalProbability(t *testing.T) {
	// Simple mortality table where death rate increases with age
	mortalityTable := make(MortalityTable, 100)
	for age := 0; age < 100; age++ {
		// Death rate increases with age (simplified)
		mortalityTable[age] = float64(age) / 1000.0
	}
	
	startAge := 30
	yearsLater := 10
	
	survivalProb := calculateSurvivalProbability(startAge, yearsLater, mortalityTable)
	
	fmt.Printf("\n=== Survival Probability Example ===\n")
	fmt.Printf("Starting age: %d\n", startAge)
	fmt.Printf("Years to survive: %d\n", yearsLater)
	fmt.Printf("Probability of surviving %d years: %.2f%%\n", 
		yearsLater, survivalProb*100)
	
	// Show year-by-year breakdown
	fmt.Println("\nYear-by-year survival chances:")
	for year := 1; year <= 5; year++ {
		prob := calculateSurvivalProbability(startAge, year, mortalityTable)
		fmt.Printf("  Year %d: %.2f%%\n", year, prob*100)
	}
	
	if survivalProb < 0 || survivalProb > 1 {
		t.Errorf("Survival probability must be between 0 and 1")
	}
}
