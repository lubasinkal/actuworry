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

// MortalityTable represents a slice of mortality rates (qx) indexed by age.
type MortalityTable []float64

// PolicyHolder represents the input parameters for an insurance policy.
type PolicyHolder struct {
	Age          int     `json:"age"`
	Term         int     `json:"term"`
	SumAssured   float64 `json:"sum_assured"`
	InterestRate float64 `json:"interest_rate"`
	TableName    string  `json:"table_name"` // e.g., "male", "female"
}

// CalculationResult holds the output of the actuarial calculations.
type CalculationResult struct {
	NetPremium      float64   `json:"net_premium"`
	ReserveSchedule []float64 `json:"reserve_schedule"`
}

// LoadMortalityTable reads a mortality table from a CSV file into a MortalityTable slice.
// It expects the CSV to have a header row, be tab-delimited, and have the qx value
// in the third column.
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
			valStr := strings.TrimSpace(rec[2])
			qx, err := strconv.ParseFloat(valStr, 64)
			if err != nil {
				valStr = strings.TrimSpace(rec[1])
				qx, err = strconv.ParseFloat(valStr, 64)
				if err != nil {
					continue
				}
			}
			table = append(table, qx)
		}
	}

	return table, nil
}

// PresentValue calculates the present value of a single future payment.
func PresentValue(amount, interestRate float64, years int) float64 {
	return amount / math.Pow(1+interestRate, float64(years))
}

// NetPremium calculates the net premium for a term life insurance policy.
// It is calculated based on the equivalence principle, where the present value
// of expected future premiums equals the present value of the expected future death benefit.
func NetPremium(p *PolicyHolder, table MortalityTable) float64 {
	var expectedFutureDeathBenefit float64
	var expectedFuturePremiums float64

	for t := 0; t < p.Term; t++ {
		// Probability of surviving to year t
		px := 1.0
		for i := 0; i < t; i++ {
			px *= (1 - table[p.Age+i])
		}

		// Probability of dying in year t
		qx := table[p.Age+t]

		// PV of death benefit paid at the end of year t+1
		deathBenefit := PresentValue(p.SumAssured, p.InterestRate, t+1)
		expectedFutureDeathBenefit += px * qx * deathBenefit

		// PV of premium paid at the beginning of year t
		premiumAnnuity := PresentValue(1, p.InterestRate, t)
		expectedFuturePremiums += px * premiumAnnuity
	}

	if expectedFuturePremiums == 0 {
		return 0
	}

	return expectedFutureDeathBenefit / expectedFuturePremiums
}

// NetPremiumReserves calculates the net premium reserve at the end of each year.
// The reserve at time t is the expected present value of future benefits minus the
// expected present value of future net premiums at that time.
func NetPremiumReserves(p *PolicyHolder, table MortalityTable, netPremium float64) []float64 {
	// The reserve schedule has n+1 elements, from t=0 to t=n.
	reserves := make([]float64, p.Term+1)

	for t := 0; t <= p.Term; t++ {
		// At the end of the term (t=n), the reserve is 0.
		if t == p.Term {
			reserves[t] = 0
			continue
		}

		var futureDeathBenefit float64
		var futurePremiums float64

		// Calculate the reserve for a policy of age x+t with remaining term n-t.
		for i := 0; i < p.Term-t; i++ {
			// Probability of surviving from age x+t to age x+t+i
			px := 1.0
			for j := 0; j < i; j++ {
				px *= (1 - table[p.Age+t+j])
			}

			// Probability of dying in the following year
			qx := table[p.Age+t+i]

			deathBenefit := PresentValue(p.SumAssured, p.InterestRate, i+1)
			futureDeathBenefit += px * qx * deathBenefit

			premiumAnnuity := PresentValue(1, p.InterestRate, i)
			futurePremiums += px * premiumAnnuity
		}

		reserves[t] = futureDeathBenefit - (netPremium * futurePremiums)
	}

	return reserves
}