# 🏗️ Actuworry - System Architecture

## Overview
Actuworry is a production-ready actuarial platform built with Go backend and vanilla JavaScript frontend. The system follows Go best practices with a modular, maintainable structure and is deployed on Render.

🌐 **Live URL:** https://actuworry.onrender.com

## 🗂️ Project Structure

```
actuworry/
├── backend/
│   ├── actuarial/
│   │   ├── core.go            # Core actuarial calculations
│   │   └── core_test.go       # Actuarial tests
│   ├── cmd/
│   │   └── server/
│   │       └── main.go         # Server entry point
│   ├── data/
│   │   ├── male.csv           # Male mortality table
│   │   └── female.csv         # Female mortality table
│   ├── handlers/
│   │   └── actuarial_handlers.go  # HTTP request handlers
│   ├── middleware/
│   │   └── cors.go            # Middleware (CORS, logging)
│   ├── models/
│   │   └── policy.go          # Data models and types
│   ├── routes/
│   │   └── routes.go          # Route configuration
│   ├── services/
│   │   └── actuarial_service.go   # Business logic layer
│   ├── scripts/              # Utility scripts
│   │   └── run.sh            # Development run script
│   ├── tests/                # Test files
│   │   ├── test_api.sh       # API test scripts
│   │   └── *.json            # Test data files
│   └── utils/                # Helper functions
│
├── frontend/
│   ├── index.html            # Main application UI
│   ├── components/           # UI components
│   ├── css/                  # Stylesheets
│   └── js/                   # JavaScript modules
│
├── docs/
│   └── ARCHITECTURE.md       # This file
├── main.go                   # Root entry point for deployment
├── go.mod                    # Go module definition
├── Makefile                  # Build automation
├── render.yaml               # Render deployment config
└── README.md                 # Project documentation
```

## 🔧 Backend Architecture

### Layered Architecture
1. **Handlers Layer** (`handlers/`)
   - HTTP request/response handling
   - Input validation
   - Error responses

2. **Service Layer** (`services/`)
   - Business logic
   - Data transformation
   - Orchestration

3. **Model Layer** (`models/`)
   - Data structures
   - Request/Response types
   - Validation rules

4. **Core Domain** (`actuarial/`)
   - Pure actuarial calculations
   - No external dependencies
   - Testable functions

### API Endpoints (Updated)
All endpoints now use `/api` prefix for clarity:

- `GET  /api/health` - Health check with service status
- `GET  /api/tables` - List available mortality tables
- `POST /api/calculate` - Single premium calculation
- `POST /api/calculate/batch` - Batch calculations
- `POST /api/calculate/sensitivity` - Sensitivity analysis
- `POST /api/analyze/portfolio` - Portfolio analysis

## 🎨 Frontend Architecture

### Technologies
- **Alpine.js** - Reactive UI without build step
- **Tailwind CSS** - Utility-first styling
- **Chart.js** - Interactive data visualization

### Features
1. **Premium Calculator**
   - Real-time calculations
   - Product type selection
   - Underwriting factors
   - Visual charts

2. **Portfolio Analysis**
   - Multiple policy management
   - Risk distribution
   - Product distribution charts

3. **Sensitivity Analysis**
   - Interest rate sensitivity
   - Age impact analysis
   - Coverage amount variations

## 🚀 Running the Application

### Development Mode
```bash
# Using Makefile
make run

# Or directly with Go
go run main.go

# Or using the backend server directly
go run backend/cmd/server/main.go
```

### Production Build
```bash
# Build using Makefile
make build

# Or build manually
go build -tags netgo -ldflags '-s -w' -o app

# Run the binary
./app
```

### Environment Variables
```bash
PORT=8080  # Server port (default: 8080)
```

## 🌐 Deployment

### Render Deployment
The application is deployed on Render using the `render.yaml` configuration:

```yaml
services:
  - type: web
    name: actuworry
    runtime: go
    buildCommand: go build -tags netgo -ldflags '-s -w' -o app
    startCommand: ./app
    envVars:
      - key: PORT
        value: 10000
```

### Deployment Process
1. Push changes to GitHub main branch
2. Render automatically detects changes
3. Builds the Go application
4. Deploys to production
5. Available at https://actuworry.onrender.com

### Manual Deployment
```bash
# Deploy using Makefile
make deploy

# Or manually
git add .
git commit -m "Deploy updates"
git push origin main
```

## 📊 Key Improvements

### Backend
✅ **Modular Structure** - Clear separation of concerns
✅ **Service Layer** - Business logic isolated from HTTP
✅ **Better Error Handling** - Consistent error responses
✅ **Middleware Support** - CORS, logging, extensible
✅ **Testable Code** - Services can be unit tested

### Frontend
✅ **Reactive UI** - Alpine.js for interactivity
✅ **Professional Design** - Tailwind CSS styling
✅ **Interactive Charts** - Multiple chart types
✅ **Tab Navigation** - Better UX organization
✅ **Form Validation** - Client-side validation

## 🧪 Testing

### Backend Tests
```bash
# Run actuarial tests
go test ./backend/actuarial -v

# Run all tests
go test ./... -v

# Run with coverage
go test ./... -cover
```

### API Testing
```bash
# Test single calculation
curl -X POST http://localhost:8080/api/calculate \
  -H "Content-Type: application/json" \
  -d '{
    "age": 35,
    "term": 10,
    "sum_assured": 100000,
    "interest_rate": 0.05,
    "table_name": "male",
    "product_type": "term_life"
  }'

# Test health endpoint
curl http://localhost:8080/api/health
```

## 📝 Development Guidelines

### Adding New Features
1. Define models in `models/`
2. Add business logic to `services/`
3. Create handlers in `handlers/`
4. Register routes in `routes/`
5. Update frontend in `js/app.js`

### Code Style
- Use descriptive variable names
- Keep functions small and focused
- Add comments for complex logic
- Follow Go conventions

## 🔄 Migration from Old Structure

The refactored code maintains backward compatibility:
- Old API endpoints still work (redirected internally)
- Frontend can be switched by renaming files
- Data structures remain compatible

To use the new frontend:
```bash
mv frontend/index.html frontend/index_old.html
mv frontend/index_new.html frontend/index.html
```

## 📚 Dependencies

### Backend
- Standard Go library (no external deps needed)
- Optional: Add database driver for persistence

### Frontend
- Alpine.js (CDN)
- Tailwind CSS (CDN)
- Chart.js (CDN)

## 🎯 Future Enhancements

### Short Term
- [ ] Add input validation middleware
- [ ] Implement caching layer
- [ ] Add API documentation (Swagger)
- [ ] Create Docker container

### Long Term
- [ ] Database persistence
- [ ] User authentication
- [ ] Real-time updates (WebSocket)
- [ ] Export functionality (PDF/Excel)

## 🤝 Contributing

1. Follow the modular structure
2. Write tests for new features
3. Update documentation
4. Use meaningful commit messages

---

Built with ❤️ for the Botswana insurance market
