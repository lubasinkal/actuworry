# Actuworry Frontend

Static frontend served directly by the Go server. No build step.

## Structure

```
frontend/
├── css/style.css   # Design system + layout
├── js/app.js      # Alpine component and Chart.js renderers
└── index.html     # App shell and calculator panels
```

## Development

```bash
make run
```

Open http://localhost:8080. The UI calls the API under `/api`.
