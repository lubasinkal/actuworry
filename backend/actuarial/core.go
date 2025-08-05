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

// MortalityTable represents a slice of qx values.
type MortalityTable []float64

// PolicyHolder represents the user input for the calculations.
type PolicyHolder struct {
	Age          int     `json:"age"`
	Term         int     `json:"term"`
	SumAssured   float64 `json:"sum_assured"`
	InterestRate float64 `json:"interest_rate"`
}

// CalculationResult represents the results of the actuarial calculations.
type CalculationResult struct {
	NetPremium      float64   `json:"net_premium"`
	ReserveSchedule []float64 `json:"reserve_schedule"`
}

// LoadMortalityTable reads a mortality table from a CSV file.
func LoadMortalityTable(path string) (MortalityTable, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1 // Allow variable number of fields
	reader.Comma = '\t'      // Use tab as delimiter

	// Skip header
	if _, err := reader.Read(); err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	var table MortalityTable
	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read record: %w", err)
		}

		if len(rec) > 2 {
			// The qx values are in the third column, but some rows have a different format.
			// We'll try to parse the third column first, and if that fails, we'll try the second.
			valStr := strings.TrimSpace(rec[2])
			qx, err := strconv.ParseFloat(valStr, 64)
			if err != nil {
				valStr = strings.TrimSpace(rec[1])
				qx, err = strconv.ParseFloat(valStr, 64)
				if err != nil {
					continue // Skip rows that don't have a valid qx value.
				}
			}
			table = append(table, qx)
		}
	}

	return table, nil
}

// PresentValue calculates the present value of a future payment.
func PresentValue(amount, interestRate float64, years int) float64 {
	return amount / math.Pow(1+interestRate, float64(years))
}

// NetPremium calculates the net premium for a given policy.
func NetPremium(p *PolicyHolder, table MortalityTable) float64 {
	var expectedFutureDeathBenefit float64
	var expectedFuturePremiums float64

	for t := 0; t < p.Term; t++ {
		px := 1.0
		for i := 0; i < t; i++ {
			px *= (1 - table[p.Age+i])
		}
		qx := table[p.Age+t]
		deathBenefit := PresentValue(p.SumAssured, p.InterestRate, t+1)
		expectedFutureDeathBenefit += px * qx * deathBenefit

		premiumAnnuity := PresentValue(1, p.InterestRate, t)
		expectedFuturePremiums += px * premiumAnnuity
	}

	return expectedFutureDeathBenefit / expectedFuturePremiums
}

// NetPremiumReserves calculates the net premium reserves for each policy year.
func NetPremiumReserves(p *PolicyHolder, table MortalityTable, netPremium float64) []float64 {
	reserves := make([]float64, p.Term)

	for t := 0; t < p.Term; t++ {
		var futureDeathBenefit float64
		var futurePremiums float64

		for i := t; i < p.Term; i++ {
			px := 1.0
			for j := t; j < i; j++ {
				px *= (1 - table[p.Age+j])
			}
			qx := table[p.Age+i]
			deathBenefit := PresentValue(p.SumAssured, p.InterestRate, i-t+1)
			futureDeathBenefit += px * qx * deathBenefit

			premiumAnnuity := PresentValue(1, p.InterestRate, i-t)
			futurePremiums += px * premiumAnnuity
		}

		reserves[t] = futureDeathBenefit - (netPremium * futurePremiums)
	}

	return reserves
}
