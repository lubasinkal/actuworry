package actuarial

import (
	"fmt"
	"math"

	"github.com/lubasinkal/v-star/pkg/rates"
	"github.com/lubasinkal/v-star/pkg/risk"
	"github.com/lubasinkal/v-star/pkg/stochastic"
)

type RiskReport struct {
	Mean   float64 `json:"mean"`
	StdDev float64 `json:"std_dev"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	VaR95  float64 `json:"var_95"`
	VaR99  float64 `json:"var_99"`
	CTE95  float64 `json:"cte_95"`
	CTE99  float64 `json:"cte_99"`
}

type VStarRateConverter struct {
	Effective float64
	inner     *rates.RateConverter
}

func NewVStarConverter(r float64) *VStarRateConverter {
	return &VStarRateConverter{
		Effective: r,
		inner:     rates.NewRateConverter(r),
	}
}

func (v *VStarRateConverter) PresentValue(amt float64, term int) float64 {
	return v.inner.PresentValue(amt, term)
}

func (v *VStarRateConverter) DiscountFactor(term int) float64 {
	return v.inner.Discount(term)
}

func (v *VStarRateConverter) V() float64 {
	return v.inner.V()
}

func (v *VStarRateConverter) VStar(j float64) float64 {
	return v.inner.VStar(j)
}

func (v *VStarRateConverter) AnnuityImmediate(n int) float64 {
	return rates.AnnuityCertainImmediate(v.Effective, n)
}

func (v *VStarRateConverter) AnnuityDue(n int) float64 {
	return rates.AnnuityCertainDue(v.Effective, n)
}

func NominalToEffective(nom float64, m int) float64 {
	return rates.NominalToEffective(nom, m)
}

func EffectiveToNominal(eff float64, m int) float64 {
	return rates.EffectiveToNominal(eff, m)
}

func ForceOfInterest(i float64) float64 {
	return rates.ForceOfInterest(i)
}

func ComputeDuration(cfs []float64, rate float64) (mac, mod, conv float64) {
	mac = rates.MacaulayDuration(rate, cfs)
	mod = rates.ModifiedDuration(rate, cfs)
	conv = rates.Convexity(rate, cfs)
	return
}

type MonteCarloEngine struct {
	gen   *stochastic.RateGenerator
	drift float64
	vol   float64
}

func NewMonteCarlo(drift, vol float64) *MonteCarloEngine {
	return &MonteCarloEngine{
		gen:   stochastic.NewRateGenerator(0.05, drift, vol),
		drift: drift,
		vol:   vol,
	}
}

func (m *MonteCarloEngine) GeneratePath(steps int, dt float64) []float64 {
	return m.gen.GeneratePath(steps, dt)
}

func (m *MonteCarloEngine) RunSimulation(numPaths, steps int, dt float64) [][]float64 {
	paths := make([][]float64, numPaths)
	for i := 0; i < numPaths; i++ {
		paths[i] = m.gen.GeneratePath(steps, dt)
	}
	return paths
}

func (m *MonteCarloEngine) RunWithSeed(numPaths, steps int, dt float64, seed uint64) [][]float64 {
	gen := stochastic.NewRateGeneratorWithSeed(0.05, m.drift, m.vol, seed)
	paths := make([][]float64, numPaths)
	for i := 0; i < numPaths; i++ {
		paths[i] = gen.GeneratePath(steps, dt)
	}
	return paths
}

func ComputeRiskReport(losses []float64) RiskReport {
	if len(losses) == 0 {
		return RiskReport{}
	}

	sum, minVal, maxVal := losses[0], losses[0], losses[0]
	for _, l := range losses {
		sum += l
		if l < minVal {
			minVal = l
		}
		if l > maxVal {
			maxVal = l
		}
	}
	mean := sum / float64(len(losses))

	var variance float64
	for _, l := range losses {
		diff := l - mean
		variance += diff * diff
	}
	stdDev := math.Sqrt(variance / float64(len(losses)))

	r := risk.ComputeReport(losses)

	return RiskReport{
		Mean:   mean,
		StdDev: stdDev,
		Min:    minVal,
		Max:    maxVal,
		VaR95:  r.VaR95,
		VaR99:  r.VaR99,
		CTE95:  r.CTE95,
		CTE99:  r.CTE99,
	}
}

func ComputeVaR(losses []float64, conf float64) float64 {
	return risk.VaR(losses, conf)
}

func ComputeCTE(losses []float64, conf float64) float64 {
	return risk.CTE(losses, conf)
}

func RunMCWithRisk(numPaths int, notional, drift, vol float64, seed uint64) RiskReport {
	dt := 1.0
	steps := 10

	mc := NewMonteCarlo(drift, vol)
	var paths [][]float64
	if seed > 0 {
		paths = mc.RunWithSeed(numPaths, steps, dt, seed)
	} else {
		paths = mc.RunSimulation(numPaths, steps, dt)
	}

	losses := make([]float64, numPaths)
	for i, path := range paths {
		finalRate := path[steps-1]
		losses[i] = math.Max(0, mc.drift-finalRate) * notional
	}

	return ComputeRiskReport(losses)
}

type BondValuation struct {
	Price       float64 `json:"price"`
	MacaulayDur float64 `json:"macaulay_duration"`
	ModifiedDur float64 `json:"modified_duration"`
	Convexity   float64 `json:"convexity"`
}

func ValueBond(face, couponRate float64, years int, ytm float64) BondValuation {
	cfs := make([]float64, years)
	for i := 0; i < years-1; i++ {
		cfs[i] = face * couponRate
	}
	cfs[years-1] = face*couponRate + face

	macDur := rates.MacaulayDuration(ytm, cfs)
	modDur := rates.ModifiedDuration(ytm, cfs)
	conv := rates.Convexity(ytm, cfs)

	price := 0.0
	rc := rates.NewRateConverter(ytm)
	for t, cf := range cfs {
		price += cf * rc.PresentValue(1.0, t+1)
	}

	return BondValuation{
		Price:       price * couponRate,
		MacaulayDur: macDur,
		ModifiedDur: modDur,
		Convexity:   conv,
	}
}

type EndowmentResult struct {
	NetSinglePremium float64 `json:"net_single_premium"`
	PremiumProfit    float64 `json:"premium_with_profit"`
}

func CalcEndowmentNSP(age, term int, sa float64, mt MortalityTable, rate float64) EndowmentResult {
	rc := NewVStarConverter(rate)

	alive := 1.0
	for a := age; a < age+term-1 && a < len(mt); a++ {
		alive *= (1.0 - mt[a])
	}

	benefitPV, premiumsPV := 0.0, 0.0

	for y := 0; y < term; y++ {
		curAge := age + y
		if curAge >= len(mt) {
			break
		}

		dthPr := mt[curAge]
		survPr := 1.0
		for a := age; a < curAge; a++ {
			survPr *= (1.0 - mt[a])
		}

		df := rc.PresentValue(1.0, y+1)
		benefitPV += survPr * dthPr * sa * df
		premiumsPV += survPr * df
	}

	endowPV := rc.PresentValue(sa, term) * alive
	benefitPV += endowPV

	nsp := 0.0
	if premiumsPV > 0 {
		nsp = benefitPV / premiumsPV
	}

	return EndowmentResult{
		NetSinglePremium: math.Round(nsp*100) / 100,
		PremiumProfit:    math.Round(nsp*1.15*100) / 100,
	}
}

type RetroReserve struct {
	AccumPremiums float64 `json:"accumulated_premiums"`
	AccumClaims   float64 `json:"accumulated_claims"`
	Reserve       float64 `json:"reserve"`
}

func CalcRetrospectiveReserve(p *Policy, mt MortalityTable, np float64) RetroReserve {
	rate := p.InterestRate

	accumPrem, accumClaim := 0.0, 0.0
	alive := 1.0

	for y := 0; y < p.Term; y++ {
		ageNow := p.Age + y
		if ageNow >= len(mt) {
			break
		}

		accumPrem = (accumPrem + np) * (1 + rate)

		dthPr := mt[ageNow]
		accumClaim = (accumClaim + alive*dthPr*p.CoverageAmount) * (1 + rate)

		alive *= (1.0 - dthPr)
	}

	res := accumPrem - accumClaim
	if res < 0 {
		res = 0
	}

	return RetroReserve{
		AccumPremiums: math.Round(accumPrem*100) / 100,
		AccumClaims:   math.Round(accumClaim*100) / 100,
		Reserve:       math.Round(res*100) / 100,
	}
}

func RunMCRates(numPaths int, initRate, drift, vol float64, seed uint64) []float64 {
	dt := 1.0
	steps := 10

	var gen *stochastic.RateGenerator
	if seed > 0 {
		gen = stochastic.NewRateGeneratorWithSeed(initRate, drift, vol, seed)
	} else {
		gen = stochastic.NewRateGenerator(initRate, drift, vol)
	}

	finalRates := make([]float64, numPaths)
	for i := 0; i < numPaths; i++ {
		path := gen.GeneratePath(steps, dt)
		finalRates[i] = path[steps-1]
	}

	return finalRates
}

func BenchmarkValuation(file string) (int, float64, error) {
	fmt.Printf("Benchmark: %s with i=0.05\n", file)
	fmt.Printf("Use v-star reader pkg for full CSV streaming\n")
	return 0, 0, nil
}
