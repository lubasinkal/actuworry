# 🇧🇼 Actuworry - Life Insurance Actuarial Tool

> A locally-built life insurance pricing, valuation and reserving tool for the Botswana market

Actuworry is a working prototype of an in-house actuarial system that handles core life insurance calculations - traditionally outsourced to foreign providers. Built to demonstrate that local actuarial technology is feasible, affordable, and aligned with market needs.

---

## 🎯 Project Vision

Build a **transparent and customizable actuarial engine** that:

- 📊 **Prices life insurance products** using proper actuarial principles
- 📉 **Calculates reserves** using prospective methods
- 💰 **Includes expense loadings** and profit margins for gross premiums
- 🌍 **Uses local mortality tables** (Botswana-specific data)
- 🔧 **Is modular and extensible** for different product types
- 💼 **Reduces reliance** on expensive black-box foreign tools

---

## ✅ Current Features

### Backend (Go)
- 🧮 **Net premium calculation** using equivalence principle
- 💸 **Gross premium calculation** with configurable expenses  
- 📊 **Net premium reserves** calculation over policy term
- 🏠 **Whole life insurance** calculations with lifetime coverage
- 📋 **Mortality table loading** from CSV files (male/female tables)
- 🌐 **RESTful API** with proper error handling and validation
- 🚀 **Batch calculation API** for processing multiple policies
- 📈 **Portfolio analysis** with summary statistics
- 🧪 **Test suite** ensuring actuarial accuracy

### Frontend (HTML/JavaScript)
- 📝 **Life insurance pricing form** with product type selection
- 🏠 **Whole life insurance** support with premium paying period
- 📊 **Interactive charts** showing reserve schedules
- 💰 **Premium results display** (net vs gross)
- 📈 **Visual reserve projections** with Chart.js integration
- 💰 **Expense assumption breakdown**
- 📋 **Reserve schedule table** showing year-by-year values
- 📱 **Responsive design** using Tailwind CSS

---

## 🚀 Quick Start

### Prerequisites
- Go 1.19+ installed
- Web browser for the frontend

### Running the Application

1. **Clone and navigate to the project:**
   ```bash
   git clone https://github.com/lubasinkal/actuworry
   cd actuworry
   ```

2. **Start the server:**
   ```bash
   ./run.sh
   # or manually:
   go run backend/main.go
   ```

3. **Open your browser:**
   ```
   http://localhost:8080
   ```

4. **Try the API directly:**
   ```bash
   curl -X POST http://localhost:8080/calculate \
     -H "Content-Type: application/json" \
     -d '{
       "age": 35,
       "term": 10, 
       "sum_assured": 100000,
       "interest_rate": 0.05,
       "table_name": "female"
     }'
   ```

### Expected Response
```json
{
  "net_premium": 456.78,
  "gross_premium": 589.32,
  "reserve_schedule": [0, 123.45, 234.56, ...],
  "product_type": "term_life",
  "expenses": {
    "initial_expense_rate": 0.03,
    "renewal_expense_rate": 0.05,
    "maintenance_expense": 50.0,
    "profit_margin": 0.15
  }
}
```

## 🔧 Technical Architecture

### Backend Stack
- **Language:** Go (for performance and simplicity)
- **HTTP Server:** Standard library with custom middleware
- **Data Format:** CSV mortality tables, JSON API responses
- **Validation:** Input validation with proper error handling
- **Testing:** Go testing framework with actuarial test cases

### Actuarial Methods
- **Net Premiums:** Calculated using equivalence principle (PV benefits = PV premiums)
- **Gross Premiums:** Iterative calculation including expense loadings
- **Reserves:** Prospective method (PV future benefits - PV future premiums)
- **Mortality Tables:** Standard life table format with qx probabilities

### API Endpoints
- `POST /calculate` - Calculate premiums and reserves (single policy)
- `POST /calculate/batch` - Calculate multiple policies with summary
- `GET /tables` - List available mortality tables
- `GET /health` - Health check endpoint
- `GET /` - Serve frontend application

### New Product Types
- **Term Life Insurance** - Coverage for specified term only
- **Whole Life Insurance** - Lifetime coverage with flexible premium paying periods

---

## 🧪 Testing

**Run the actuarial tests:**
```bash
cd backend && go test -v ./actuarial/
```

**Test the API manually:**
```bash
# Health check
curl http://localhost:8080/health

# List available tables
curl http://localhost:8080/tables

# Calculate premiums
curl -X POST http://localhost:8080/calculate \
  -H "Content-Type: application/json" \
  -d @request.json
```

---

## 🎨 Customization

### Expense Assumptions
Modify the default expense structure in `backend/actuarial/core.go`:
```go
func DefaultExpenseStructure() ExpenseStructure {
    return ExpenseStructure{
        InitialExpenseRate: 0.03,   // 3% of sum assured
        RenewalExpenseRate: 0.05,   // 5% of gross premium 
        MaintenanceExpense: 50.0,   // BWP 50 per year
        ProfitMargin:      0.15,    // 15% profit margin
    }
}
```

### Adding New Mortality Tables
1. Add CSV file to `backend/data/` directory
2. Update `tablesToLoad` slice in `backend/main.go`
3. Restart server to load new table

---

## 🚧 Future Enhancements

### Short Term
- [x] **Product Types:** Whole life insurance ✅
- [x] **Batch Processing:** Multiple policy calculations ✅
- [x] **Interactive Charts:** Reserve visualization ✅
- [ ] **Product Types:** Endowments, annuities
- [ ] **Underwriting:** Risk factors, medical loadings
- [ ] **Currency:** Multi-currency support
- [ ] **Export:** PDF quotes, Excel reserve schedules

### Medium Term  
- [ ] **Database:** PostgreSQL for mortality tables and policies
- [ ] **Authentication:** User management and API keys
- [ ] **Validation:** Real-time form validation
- [ ] **Charts:** Interactive premium and reserve visualizations

### Long Term
- [ ] **Stochastic Models:** Monte Carlo simulations
- [ ] **Economic Scenarios:** Interest rate and inflation modeling
- [ ] **Regulatory:** IFRS 17, Solvency II compliance
- [ ] **Integration:** Core insurance system APIs

---

## 📚 Actuarial Background

This tool implements standard life insurance actuarial methods:

**Net Premium Calculation:**
```
Net Premium = PV(Death Benefits) / PV(Premium Annuity)
```

**Reserve Calculation (Prospective Method):**
```
Reserve at time t = PV(Future Benefits) - PV(Future Net Premiums)
```

**Gross Premium:** Net premium plus expense loadings and profit margin

For more details, see standard actuarial texts like Bowers et al. or Dickson et al.

---

## 📊 Example Calculation

**Input:**
- Male, age 35
- 10-year term life policy
- BWP 100,000 sum assured  
- 5% interest rate

**Output:**
- **Net Premium:** BWP 456.78 (pure risk premium)
- **Gross Premium:** BWP 589.32 (includes expenses and profit)
- **Reserves:** Calculated for each policy year

---

