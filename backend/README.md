# Actuworry Backend

Go-based actuarial calculations API server.

## Structure

```
backend/
├── actuarial/      # Core actuarial calculation logic
├── cmd/
│   └── server/     # Main application entry point
├── data/           # Mortality tables (CSV files)
├── handlers/       # HTTP request handlers
├── middleware/     # HTTP middleware (CORS, etc.)
├── models/         # Data models and types
├── routes/         # API route definitions
├── services/       # Business logic services
├── tests/          # Test files and scripts
├── scripts/        # Utility scripts
└── utils/          # Helper functions
```

## Running Locally

```bash
# From project root
go run ./backend/cmd/server

# Or build and run
go build -o app ./backend/cmd/server
./app
```

## Testing

```bash
# Run tests
cd backend/tests
./test_api.sh
./test_enhanced_api.sh
```

## API Endpoints

- `GET /health` - Health check
- `GET /tables` - List available mortality tables
- `POST /calculate` - Single premium calculation
- `POST /calculate/batch` - Batch premium calculations
- `POST /calculate/sensitivity` - Sensitivity analysis
- `POST /analyze/portfolio` - Portfolio analysis

## Environment Variables

- `PORT` - Server port (default: 8080)
