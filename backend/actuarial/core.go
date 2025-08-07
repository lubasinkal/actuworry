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
	Age          int     `json:"age"`
	Term         int     `json:"term"`
	CoverageAmount   float64 `json:"sum_assured"`
	InterestRate float64 `json:"interest_rate"`
	Gender    string  `json:"table_name"`
}

type PremiumCalculation struct {
	NetPremium      float64            `json:"net_premium"`
	GrossPremium    float64            `json:"gross_premium"`
	ReserveSchedule []float64          `json:"reserve_schedule"`
	ProductType     string             `json:"product_type"`
	ExpenseDetails  map[string]float64 `json:"expenses,omitempty"`
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

func CalculateFullPremium(policy *Policy, mortalityTable MortalityTable) PremiumCalculation {
	netPremium := CalculateNetPremium(policy, mortalityTable)
	expenseAssumptions := CreateDefaultExpenses()
	grossPremium := CalculateGrossPremium(policy, mortalityTable, netPremium, expenseAssumptions)
	reserveSchedule := CalculateReserveSchedule(policy, mortalityTable, netPremium)

	expenseBreakdown := map[string]float64{
		"initial_expense_rate": expenseAssumptions.InitialExpenseRate,
		"renewal_expense_rate": expenseAssumptions.RenewalExpenseRate,
		"maintenance_expense":  expenseAssumptions.MaintenanceExpense,
		"profit_margin":        expenseAssumptions.ProfitMargin,
	}

	return PremiumCalculation{
		NetPremium:      netPremium,
		GrossPremium:    grossPremium,
		ReserveSchedule: reserveSchedule,
		ProductType:     "term_life",
		ExpenseDetails:  expenseBreakdown,
	}
}

