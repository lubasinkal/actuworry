# Actuworry

Life insurance pricing, valuation and reserving for the Botswana market. Go backend, static frontend, no build step.

Live demo: https://actuworry.onrender.com

## Run

Requires Go 1.26+.

```bash
make run
```

Open http://localhost:8080. `make test` runs the Go tests.

## API

All endpoints are under `/api`.

| Method | Path | Purpose |
| --- | --- | --- |
| POST | `/api/calculate` | Single policy premium and reserve schedule |
| POST | `/api/calculate/batch` | Batch premiums |
| POST | `/api/calculate/sensitivity` | Interest, age and coverage sensitivity |
| POST | `/api/analyze/portfolio` | Portfolio summary |
| GET | `/api/tables` | Loaded mortality tables |
| GET | `/api/health` | Health check |
| POST | `/api/vstar/montecarlo` | Monte Carlo loss distribution |
| POST | `/api/vstar/risk` | VaR and CTE |
| POST | `/api/vstar/rate-convert` | Nominal, effective and force of interest |
| POST | `/api/vstar/reserve-retro` | Retrospective reserve at a policy year |

Example:

```bash
curl -X POST http://localhost:8080/api/calculate \
  -H "Content-Type: application/json" \
  -d '{"age":35,"term":20,"sum_assured":100000,"interest_rate":0.05,"table_name":"male","product_type":"term_life"}'
```

## Structure

```
backend/actuarial/   calculation engine
backend/data/        mortality tables (CSV)
backend/handlers/    HTTP handlers
backend/routes/      route definitions
frontend/            index.html, css/style.css, js/app.js
main.go              server entry point
```

Premiums use the equivalence principle; gross premiums add expenses and profit. Reserves are calculated prospectively, with a retrospective check. The UI is educational and not for commercial decisions.
