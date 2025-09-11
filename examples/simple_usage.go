// Package main shows simple examples of how to use the actuarial calculator.
// This file is meant for learning - it shows the basic flow of calculations.
package main

import (
	"actuworry/backend/actuarial"
	"actuworry/backend/models"
	"actuworry/backend/services"
	"fmt"
)

func main() {
	fmt.Println("=== Actuarial Calculator Examples ===\n")
	
	// Example 1: Basic Present Value
	showPresentValueExample()
	
	// Example 2: Simple Premium Calculation
	showSimplePremiumCalculation()
	
	// Example 3: Using the Service Layer
	showServiceLayerExample()
}

// showPresentValueExample demonstrates the time value of money
func showPresentValueExample() {
	fmt.Println("1. PRESENT VALUE EXAMPLE")
	fmt.Println("------------------------")
	fmt.Println("Question: If you need $100,000 in 10 years, how much is that worth today?")
	fmt.Println("(Assuming 5% annual interest rate)\n")
	
	futureValue := 100000.0
	interestRate := 0.05
	years := 10
	
	presentValue := actuarial.CalculatePresentValue(futureValue, interestRate, years)
	
	fmt.Printf("Future value: $%.2f\n", futureValue)
	fmt.Printf("Interest rate: %.1f%%\n", interestRate*100)
	fmt.Printf("Years: %d\n", years)
	fmt.Printf("Present value: $%.2f\n\n", presentValue)
	fmt.Printf("This means $%.2f today will grow to $%.2f in %d years at %.1f%% interest.\n\n", 
		presentValue, futureValue, years, interestRate*100)
}

// showSimplePremiumCalculation shows a basic life insurance calculation
func showSimplePremiumCalculation() {
	fmt.Println("2. SIMPLE PREMIUM CALCULATION")
	fmt.Println("-----------------------------")
	fmt.Println("Calculating premium for a 35-year-old male")
	fmt.Println("$500,000 coverage, 20-year term life insurance\n")
	
	// Create a simple mortality table (fake data for example)
	mortalityTable := createSampleMortalityTable()
	
	// Define the policy
	policy := actuarial.Policy{
		Age:            35,
		Term:           20,
		CoverageAmount: 500000,
		InterestRate:   0.04, // 4% discount rate
		Gender:         "male",
		ProductType:    "term_life",
	}
	
	// Calculate net premium (pure cost)
	netPremium := actuarial.CalculateNetPremium(&policy, mortalityTable)
	
	// Add expenses to get gross premium (what customer pays)
	expenses := actuarial.CreateDefaultExpenses()
	grossPremium := actuarial.CalculateGrossPremium(&policy, mortalityTable, netPremium, expenses)
	
	fmt.Printf("Policy Details:\n")
	fmt.Printf("  Age: %d\n", policy.Age)
	fmt.Printf("  Coverage: $%.0f\n", policy.CoverageAmount)
	fmt.Printf("  Term: %d years\n\n", policy.Term)
	
	fmt.Printf("Premium Calculation:\n")
	fmt.Printf("  Net Premium (pure cost): $%.2f/year\n", netPremium)
	fmt.Printf("  Gross Premium (with expenses): $%.2f/year\n", grossPremium)
	fmt.Printf("  Monthly Payment: $%.2f\n\n", grossPremium/12)
}

// showServiceLayerExample demonstrates using the higher-level service
func showServiceLayerExample() {
	fmt.Println("3. USING THE SERVICE LAYER")
	fmt.Println("--------------------------")
	fmt.Println("The service layer makes it easier to use the calculator\n")
	
	// Create the service
	_ = services.NewActuarialService() // In real code: service := ...
	
	// Load mortality tables (you would load real files here)
	// For this example, we'll just use fake data
	fmt.Println("Note: In real usage, you'd load actual mortality tables from CSV files")
	fmt.Println("Example: service.LoadMortalityTable(\"male\", \"data/male.csv\")\n")
	
	// Create a policy using the models package
	_ = models.Policy{
		Age:            40,
		Term:           15,
		CoverageAmount: 250000,
		InterestRate:   0.045,
		Gender:         "male",
		ProductType:    "term_life",
	}
	
	fmt.Printf("Policy Request:\n")
	fmt.Printf("  40-year-old male\n")
	fmt.Printf("  $250,000 coverage\n")
	fmt.Printf("  15-year term\n")
	fmt.Printf("  4.5%% interest rate\n\n")
	
	// In real usage, you would call:
	// result, err := service.CalculatePremium(&policy)
	
	fmt.Println("The service would return a PremiumCalculation with:")
	fmt.Println("  - Net premium (pure cost)")
	fmt.Println("  - Gross premium (what customer pays)")
	fmt.Println("  - Reserve schedule (money set aside each year)")
	fmt.Println("  - Expense breakdown")
	fmt.Println("  - Risk assessment details\n")
}

// createSampleMortalityTable creates fake mortality data for examples
func createSampleMortalityTable() actuarial.MortalityTable {
	// This creates a simple mortality table where death probability
	// increases with age. Real tables come from actuarial studies.
	table := make(actuarial.MortalityTable, 100)
	
	for age := 0; age < 100; age++ {
		// Very simplified: death rate increases with age
		if age < 20 {
			table[age] = 0.0005 // Very low for young people
		} else if age < 40 {
			table[age] = 0.001 + float64(age-20)*0.0001
		} else if age < 60 {
			table[age] = 0.003 + float64(age-40)*0.0005
		} else if age < 80 {
			table[age] = 0.013 + float64(age-60)*0.003
		} else {
			table[age] = 0.073 + float64(age-80)*0.01
		}
	}
	
	return table
}

// Tips for Understanding the Code:
// ================================
// 
// 1. Start with Present Value - it's the foundation
//    - Money today is worth more than money tomorrow
//    - We discount future payments to today's value
//
// 2. Mortality Tables are just lists of death probabilities
//    - Index = age, Value = probability of dying that year
//    - Real tables come from population studies
//
// 3. Premium Calculation balances two things:
//    - Expected payouts (death benefits)
//    - Expected premium collections
//    - The premium makes these equal in present value terms
//
// 4. Net vs Gross Premium:
//    - Net = pure cost of insurance (just the death risk)
//    - Gross = net + expenses + profit (what customer pays)
//
// 5. The Service Layer:
//    - Handles the HTTP API stuff
//    - Loads and manages mortality tables
//    - Converts between API models and calculation models
