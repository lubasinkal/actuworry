package actuarial

import (
	"math"
	"testing"
)

func TestBondPrice(t *testing.T) {
	result := ValueBond(1000, 0.06, 10, 0.05)

	expected := 0.0
	for year := 1; year <= 10; year++ {
		cashFlow := 60.0
		if year == 10 {
			cashFlow += 1000
		}
		expected += cashFlow / math.Pow(1.05, float64(year))
	}

	if !floatEquals(result.Price, expected, 0.01) {
		t.Errorf("Expected bond price %f, got %f", expected, result.Price)
	}
}

func TestEndowmentSinglePremium(t *testing.T) {
	table := make(MortalityTable, 101)
	table[35] = 0.002

	// A one-year endowment always pays at the end of the year, so its
	// single premium is simply the discounted sum assured.
	result := CalcEndowmentNSP(35, 1, 1000, table, 0.05)
	expected := 1000 / 1.05

	if !floatEquals(result.NetSinglePremium, expected, 0.01) {
		t.Errorf("Expected endowment NSP %f, got %f", expected, result.NetSinglePremium)
	}
}

func TestMonteCarloUsesCurrentRateAsLossThreshold(t *testing.T) {
	result := RunMCWithRisk(2000, 1000000, 0.05, 0.02, 0.15, 42)

	if result.VaR95 <= 0 {
		t.Errorf("Expected a positive 95%% VaR when the final rate can fall below 5%%, got %f", result.VaR95)
	}
}

func TestRetrospectiveReserveMatchesProspective(t *testing.T) {
	table := make(MortalityTable, 100)
	table[35] = 0.002
	table[36] = 0.003
	table[37] = 0.004
	table[38] = 0.005

	full := &Policy{
		Age:            35,
		Term:           4,
		CoverageAmount: 1000,
		InterestRate:   0.05,
		ProductType:    "term_life",
	}
	netPremium := CalculateNetPremium(full, table)
	prospective := CalculateTermLifeReserveSchedule(full, table, netPremium)

	part := *full
	part.Term = 2
	retrospective := CalcRetrospectiveReserve(&part, table, netPremium)

	if !floatEquals(retrospective.Reserve, prospective[2], 0.01) {
		t.Errorf("Expected retrospective reserve %f to match prospective %f", retrospective.Reserve, prospective[2])
	}
	if !floatEquals(retrospective.AccumPremiums-retrospective.AccumClaims, retrospective.Reserve, 0.01) {
		t.Errorf("Accumulated premiums minus claims should equal the reserve")
	}
}
