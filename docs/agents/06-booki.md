# 06 — Booki (Morph Booki)

## Overview

Booki is an accounting and booking platform with AI-powered data analysis, double-entry bookkeeping, warehouse management, and flow log tracking.

- **Module:** `github.com/academi/booki` (Go 1.24)
- **Backend:** `booki/backend/` — Go + Gin
- **Frontend:** `booki/frontend/` — React 19 + Vite 8 + TailwindCSS 4
- **Ports:** API `9095`, UI `5174`
- **Auth:** UsersPanel SSO (with local JWT exchange)

## Backend architecture

### Entry point: `booki/backend/cmd/server/main.go`

1. Load config from env
2. Connect MySQL, Redis (optional)
3. Create Gin router with CORS + auth middleware
4. Register all routes via `api.NewRouter()`
5. Listen on `APP_PORT` (default `9095`)

### Package layout

```
booki/backend/
├── cmd/server/main.go           ← Entry point
├── internal/
│   ├── api/router.go            ← Route registration
│   ├── config/config.go         ← Env loading
│   ├── handlers/                ← 10 handler files
│   │   ├── accounting.go        ← Accounts, journal entries, ledger
│   │   ├── assets.go            ← Asset management + Morph sync
│   │   ├── assistant.go         ← AI assistant (intent-based)
│   │   ├── booking.go           ← Bookings CRUD + status
│   │   ├── customers.go         ← Customer CRUD
│   │   ├── dashboard.go         ← Dashboard data
│   │   ├── flow_log.go          ← Flow log entries + AI analysis
│   │   ├── imports.go           ← CSV/JSON/HTTP/ledger imports
│   │   ├── reports.go           ← Trial balance, profit/loss
│   │   └── warehouse.go         ← Warehouse, stock, products
│   ├── models/                  ← Data structures
│   └── middleware/               ← Auth middleware
├── go.mod
└── go.sum
```

### Route groups (all under `/api/v1`)

| Group | Key endpoints |
|-------|---------------|
| Auth | `POST /auth/register`, `POST /auth/login`, `POST /auth/platform-session`, `POST /auth/refresh`, `POST /auth/logout`, `GET /auth/me` |
| Assistant | `POST /assistant/chat` — MorphAI contract (intents: create_customer, create_booking, list_*, analyze_*, calculate) |
| Organization | `GET /organization`, `PATCH /organization` |
| Dashboard | `GET /dashboard` |
| Accounting | `GET /accounts`, `POST /journal-entries`, `GET /journal-entries`, `GET /ledger` |
| Reports | `GET /reports/trial-balance`, `GET /reports/profit-loss` |
| Customers | `GET/POST /customers` |
| Bookings | `GET/POST /bookings`, `GET/PATCH/DELETE /bookings/:id`, `PATCH /bookings/:id/status`, `POST /bookings/:id/post` |
| Warehouse | `GET/POST /warehouses`, `GET/PATCH/DELETE /warehouses/:id`, `GET/POST /products`, `GET /warehouse/stock`, `GET /warehouse/movements`, `POST /warehouse/stock-in`, `POST /warehouse/stock-out`, `POST /warehouse/transfers` |
| Flow Log | `GET/POST /flow-log/entries`, `PATCH/DELETE /flow-log/entries/:id`, `GET /flow-log/summary`, `POST /flow-log/analyze` |
| Assets | `GET/POST /assets`, `POST /assets/sync-morph` |
| Imports | `POST /imports/csv`, `POST /imports/json`, `POST /imports/http`, `POST /imports/ledger`, `GET /imports/logs` |

### Database

- **MySQL:** All structured data (accounts, journal entries, customers, bookings, warehouse, products, flow log, assets)
- **Redis:** Optional caching

### Auth flow

1. Client logs in via `POST /api/v1/auth/login` → proxies to UsersPanel
2. Or uses `POST /api/v1/auth/platform-session` to exchange UsersPanel JWT for local Booki JWT
3. Local JWT used for subsequent requests
4. Dev mode: reuses shared `userspanel_session_token` cookie

### AI assistant

Intent-based assistant following MorphAI contract. Supported intents:
- `create_customer`, `create_booking`
- `list_customers`, `list_bookings`, `list_products`, `list_warehouses`
- `analyze_flow_log`, `analyze_financials`
- `calculate_profit`, `calculate_balance`

## Frontend architecture

### Build & tooling
- React 19 + TypeScript 6.0 + Vite 8 + TailwindCSS 4
- State: Zustand v5
- Routing: react-router-dom v7
- Port: 5174 (dev)

### Key pages
- `LoginPage`, `RegisterPage`
- `DataAiPage` — AI-powered data analysis
- `BookingsPage` — Booking management
- `FlowLogPage` — Flow log tracking

### Shared components
- `@robo/platform-chat` — AI assistant chat drawer
- `PlatformAssistantDrawer.tsx` — Booki-specific wrapper

## How to add a new feature

1. **Backend handler:** Add file in `booki/backend/internal/handlers/`
2. **Route:** Register in `booki/backend/internal/api/router.go`
3. **Model:** Add struct in `booki/backend/internal/models/`
4. **Frontend page:** Add in `booki/frontend/src/pages/`
5. **Route:** Add in `booki/frontend/src/App.tsx`

## Key env vars

| Variable | Default | Purpose |
|----------|---------|---------|
| `APP_PORT` | `9095` | Backend port |
| `DATABASE_URL` | — | MySQL DSN |
| `REDIS_ADDR` | — | Redis address (optional) |
| `JWT_SECRET` | — | Local JWT signing secret |
| `USERS_PANEL_BASE_URL` | — | UsersPanel URL |
| `MORPH_AI_API_KEY` | — | DashScope API key |
| `MORPH_AI_MODEL` | `qwen3-max` | AI model |