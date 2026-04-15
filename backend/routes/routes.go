package routes

import (
	"actuworry/backend/handlers"
	"actuworry/backend/middleware"
	"net/http"
)

// SetupRoutes configures all application routes
func SetupRoutes(handler *handlers.ActuarialHandler) *http.ServeMux {
	mux := http.NewServeMux()

	// Standard API routes
	mux.HandleFunc("/api/calculate",
		middleware.Chain(handler.CalculatePremium, middleware.Logger, middleware.CORS))

	mux.HandleFunc("/api/calculate/batch",
		middleware.Chain(handler.CalculateBatch, middleware.Logger, middleware.CORS))

	mux.HandleFunc("/api/calculate/sensitivity",
		middleware.Chain(handler.SensitivityAnalysis, middleware.Logger, middleware.CORS))

	mux.HandleFunc("/api/analyze/portfolio",
		middleware.Chain(handler.PortfolioAnalysis, middleware.Logger, middleware.CORS))

	mux.HandleFunc("/api/tables",
		middleware.Chain(handler.GetTables, middleware.Logger, middleware.CORS))

	mux.HandleFunc("/api/health",
		middleware.Chain(handler.HealthCheck, middleware.Logger, middleware.CORS))

	// v-star advanced features
	mux.HandleFunc("/api/vstar/montecarlo",
		middleware.Chain(handler.MonteCarloSimulation, middleware.Logger, middleware.CORS))

	mux.HandleFunc("/api/vstar/risk",
		middleware.Chain(handler.RiskAnalysis, middleware.Logger, middleware.CORS))

	mux.HandleFunc("/api/vstar/duration",
		middleware.Chain(handler.DurationCalculator, middleware.Logger, middleware.CORS))

	mux.HandleFunc("/api/vstar/rate-convert",
		middleware.Chain(handler.RateConverterHandler, middleware.Logger, middleware.CORS))

	mux.HandleFunc("/api/vstar/endowment",
		middleware.Chain(handler.EndowmentCalculator, middleware.Logger, middleware.CORS))

	mux.HandleFunc("/api/vstar/reserve-retro",
		middleware.Chain(handler.RetrospectiveReserve, middleware.Logger, middleware.CORS))

	mux.HandleFunc("/api/vstar/bond",
		middleware.Chain(handler.BondValuation, middleware.Logger, middleware.CORS))

	// Static file server for frontend
	fs := http.FileServer(http.Dir("frontend/"))
	mux.Handle("/", fs)

	return mux
}
