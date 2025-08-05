package actuarial

import (
	"math"
	"testing"
)

// A small, predictable mortality table for testing.
// The slice needs to be large enough to be indexed by age.
var testMortalityTable = make(MortalityTable, 100)

func init() {
	testMortalityTable[35] = 0.002 // qx at age 35
	testMortalityTable[36] = 0.003 // qx at age 36
	testMortalityTable[37] = 0.004 // qx at age 37
}

// Helper function to check for approximate equality of float64 values.
func floatEquals(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}

func TestNetPremium(t *testing.T) {
	policyHolder := &PolicyHolder{
		Age:          35,
		Term:         2,
		SumAssured:   1000,
		InterestRate: 0.05,
	}

	// Expected premium calculated manually for comparison.
	expectedPremium := 2.36879

	actualPremium := NetPremium(policyHolder, testMortalityTable)

	if !floatEquals(expectedPremium, actualPremium, 0.0001) {
		t.Errorf("Expected Net Premium %f, but got %f", expectedPremium, actualPremium)
	}
}

func TestNetPremiumReserves(t *testing.T) {
	policyHolder := &PolicyHolder{
		Age:          35,
		Term:         2,
		SumAssured:   1000,
		InterestRate: 0.05,
	}
	// Use the *actual* calculated premium, not a rounded one.
	netPremium := NetPremium(policyHolder, testMortalityTable)

	// Expected values calculated manually for a schedule of size n+1
	// Reserve at t=0 is always 0 (by definition of net premium)
	// Reserve at t=1 (end of year 1):
	//   PV Future Benefits at age 36: v * q_36 * SA = (1/1.05) * 0.003 * 1000 = 2.85714
	//   PV Future Premiums at age 36: 1 * Premium = 2.36879...
	//   Reserve = 2.85714 - 2.36879... = 0.48835
	// Reserve at t=2 (end of term) is always 0
	expectedReserves := []float64{0.0, 0.48835, 0.0}

	actualReserves := NetPremiumReserves(policyHolder, testMortalityTable, netPremium)

	if len(actualReserves) != policyHolder.Term+1 {
		t.Fatalf("Expected reserve schedule of length %d, but got %d", policyHolder.Term+1, len(actualReserves))
	}

	if !floatEquals(expectedReserves[0], actualReserves[0], 0.0001) {
		t.Errorf("Expected Reserve at t=0 to be %f, but got %f", expectedReserves[0], actualReserves[0])
	}
	if !floatEquals(expectedReserves[1], actualReserves[1], 0.0001) {
		t.Errorf("Expected Reserve at t=1 to be %f, but got %f", expectedReserves[1], actualReserves[1])
	}
	if !floatEquals(expectedReserves[2], actualReserves[2], 0.0001) {
		t.Errorf("Expected Reserve at t=2 to be %f, but got %f", expectedReserves[2], actualReserves[2])
	}
}