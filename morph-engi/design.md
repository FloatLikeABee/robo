# Morph Engi — Civil Engineering Project Platform

## 1. Product Vision

**Morph Engi** is a daily-work platform for civil engineers and site teams. It is **project-first**: every feature rolls up to an active construction or infrastructure job. Like **Booki**, nothing should feel like enterprise ERP—clean cards, guided forms, plain language, and an AI assistant that can do anything a user can do in the UI.

### Target users

| Persona | Primary needs |
|---------|----------------|
| Project engineer | Schedule, drawings, quantities, RFIs |
| Site manager | Crew, materials, daily logs, contractors |
| Quantity surveyor | Budget vs actual, BOQ, cost codes |
| Client liaison | Stakeholders, permits, public updates |
| Small consultancy owner | Multi-project overview, cash flow |

### Non-goals (v1)

- Replace full BIM authoring (Revit, Civil 3D)—we **link and track** models/drawings
- Replace native CAD—store metadata, versions, viewer links
- ERP-level payroll/GL—lightweight project finance only

---

## 2. Core Philosophy (Booki-aligned)

1. **Simplicity first** — one obvious next action per screen
2. **Project context everywhere** — global project switcher in header
3. **Mobile-ready** — bottom nav on phone, cards not dense grids
4. **AI as co-pilot** — natural language creates/updates/lists all entities
5. **Platform-ready** — UsersPanel auth, MorphAI assistant contract, optional Morph broadcast

---

## 3. Visual Design

Reuse Booki design tokens for familiarity:

| Token | Value | Use |
|-------|-------|-----|
| Deep Violet | `#5B3FD6` | Primary, active nav |
| Dark Grey | `#1E1E24` | Background |
| Accent Yellow | `#F5C542` | CTAs, highlights |
| Surface | `#2B2B33` | Panels |
| Engineering Teal | `#2DD4BF` | Secondary accent (Engi identity) |

**Layout:** sidebar (desktop) + top bar (project switcher, AI, messages) + glass cards + 20px radius.

**Mood:** Linear × Booki × construction dashboard—not cluttered Gantt ERP.

---

## 4. Information Architecture

```text
Morph Engi
├── Dashboard (project snapshot)
├── Projects ★ hub
│   ├── Overview & phases
│   ├── Tasks / milestones
│   └── Daily site log
├── Materials
│   ├── Catalog (cement, steel, aggregate…)
│   └── Project usage & requisitions
├── BIM & CAD
│   ├── Model register (IFC, RVT links)
│   └── Drawing sheets (DWG/PDF metadata)
├── Design Charts
│   ├── Calculation sheets (beam, footing, earthwork)
│   └── Reference curves & code checks (stored results)
├── Finance & Resources
│   ├── Budget lines (BOQ-style)
│   ├── Expenses & commitments
│   └── Equipment / labor resource pool
├── Contractors
│   ├── Vendor directory
│   └── Contracts & work orders per project
├── Workers
│   ├── Crew roster
│   └── Shift / attendance on site
├── Public Relations
│   ├── Stakeholders (client, council, community)
│   └── Communications log (meetings, permits, notices)
└── Settings (org, units, currency, roles)
```

---

## 5. Module Specifications

### 5.1 Projects (primary module)

**Purpose:** Single source of truth for job identity and progress.

| Field | Type | Notes |
|-------|------|-------|
| code | string | e.g. `BRG-2026-01` |
| name | string | Human title |
| client | string | |
| location | string | Site address / coords |
| status | enum | `planning`, `active`, `on_hold`, `complete` |
| start_date / end_date | date | |
| budget_total | decimal | Planned contract value |
| progress_pct | 0–100 | Manual or derived from tasks |
| description | text | |

**Sub-entities:**

- **Phases** — foundation, superstructure, finishing (name, target date, status)
- **Tasks** — assignable items (title, assignee, due, status, priority)
- **Site logs** — daily note (weather, crew count, summary, issues)

**UX flows:**

```text
New user → pick/create org → create first project → dashboard populated
Daily open → select project → site log → materials used → update progress
```

### 5.2 Materials

**Catalog:** name, category (concrete, steel, earthworks, MEP…), unit (m³, ton, bag), unit_cost, supplier, stock_on_hand, reorder_level.

**Usage:** link material + qty + project + date (+ optional task). Auto-compute line cost.

**Simplified UX:** “Add 40 bags cement to Bridge Project today” via AI or 3-field form.

### 5.3 BIM & CAD

Not a viewer—**document register:**

| doc_type | Examples |
|----------|----------|
| bim | IFC, RVT cloud link |
| cad | DWG, DXF path |
| drawing | PDF sheet |
| spec | Specification section |

Fields: project_id, name, version, status (draft/issued/for construction), file_url or external_id, discipline (structural, civil, arch), notes.

Future: integrate Autodesk ACC / BIM 360 links.

### 5.4 Design Charts

Engineering calculation **records** (not a full FEA solver):

- chart_type: `beam`, `column`, `footing`, `retaining_wall`, `earthwork`, `hydrology`, `custom`
- inputs_json: span, load, fc, etc.
- results_json: utilization, safety factor, pass/fail
- notes + attachment link

UI: list + detail panel; optional simple chart preview (SVG bar for utilization). AI can create/update and explain results.

### 5.5 Finance & Resources

**Budget lines:** project_id, cost_code, description, planned_amount, actual_amount, category (labor, material, equipment, subcontract, overhead).

**Resources:** equipment and labor pools—name, type, cost_per_day, availability status.

**Allocations:** resource_id + project_id + date range + qty/days.

Dashboard widgets: budget burn, variance %, top cost codes.

### 5.6 Contractors

Directory: name, trade (earthworks, concrete, steel, MEP), contact, phone, email, rating, status.

**Contracts:** project_id, contractor_id, scope, contract_value, status (draft/active/complete), start/end.

### 5.7 Workers

Roster: name, role, trade, phone, certifications (JSON list), status (active/inactive).

**Shifts:** project_id, worker_id, date, hours, notes.

### 5.8 Public Relations

**Stakeholders:** name, organization, role (client, regulator, community, media), contact, influence (low/med/high), sentiment (positive/neutral/negative).

**Communications:** stakeholder_id, project_id, channel (meeting/email/notice/permit), subject, body, occurred_at.

---

## 6. AI Assistant (Morph Engi AI)

Follows [`AI_ASSISTANT_MORPHAI_CONTRACT.md`](../AI_ASSISTANT_MORPHAI_CONTRACT.md).

### Capabilities

The assistant must perform **every functional UI action** via tools:

| Domain | Read tools | Write tools |
|--------|------------|-------------|
| Projects | list_projects, get_project, list_tasks, list_site_logs | create_project, update_project, create_task, create_site_log |
| Materials | list_materials, list_material_usage | create_material, record_material_usage |
| Documents | list_documents | create_document, update_document |
| Charts | list_design_charts | create_design_chart, update_design_chart |
| Finance | list_budget_lines, list_resources, dashboard_summary | create_budget_line, update_budget_line, allocate_resource |
| Contractors | list_contractors, list_contracts | create_contractor, create_contract |
| Workers | list_workers, list_shifts | create_worker, record_shift |
| PR | list_stakeholders, list_communications | create_stakeholder, log_communication |

### Behavior rules

1. JSON tool call when live data or writes needed
2. Multi-turn state for partial create flows (`state.intent` + `state.fields`)
3. Confirm destructive actions (delete) in markdown before executing
4. Default project context from UI `state.active_project_id`
5. Summarize lists as markdown tables; keep numbers formatted

### Example prompts

- “Create project code BRG-01 name Riverside Bridge status active budget 2.4M”
- “Log 12 tons of rebar on Riverside today”
- “Who are the stakeholders for Riverside and any negative sentiment?”
- “Show budget variance for active projects”
- “Add RFI drawing sheet S-104 version 3 for foundation”

---

## 7. Technical Architecture

```text
┌─────────────────────────────────────────────────────────┐
│  Svelte 5 + Vite + Tailwind (port 5179)                 │
│  Booki-like layout · platform-chat drawer               │
└───────────────────────────┬─────────────────────────────┘
                            │ REST /api/v1/*
┌───────────────────────────▼─────────────────────────────┐
│  Rust Axum API (port 9096)                              │
│  JWT · UsersPanel SSO · MorphAI tool loop               │
└───────────────────────────┬─────────────────────────────┘
                            │
         ┌──────────────────┼──────────────────┐
         ▼                  ▼                  ▼
    MySQL / SQLite    UsersPanel          DashScope
    morph_engi        auth/messages       MORPH_AI_*
```

### Stack

| Layer | Choice |
|-------|--------|
| Backend | Rust, Axum 0.7, SQLx, jsonwebtoken |
| Frontend | Svelte 5, Vite 7, Tailwind 4 |
| DB | MySQL 8 (prod) or SQLite (local zero-config) |
| AI | `pkg/morphai-rs` |
| Auth | UsersPanel JWT + local org/user mapping |

### API conventions

- Prefix: `/api/v1`
- Auth: `Authorization: Bearer <token>`
- Org scoping: all queries filter by `organization_id` from JWT
- Errors: `{ "error": "message" }`
- Lists: `{ "items": [...] }` or named keys (`projects`, `materials`)

---

## 8. Data Model (summary)

```text
organizations ─┬─ users
               ├─ projects ─┬─ project_phases
               │            ├─ project_tasks
               │            └─ site_logs
               ├─ materials ─ material_usages → projects
               ├─ documents → projects
               ├─ design_charts → projects
               ├─ budget_lines → projects
               ├─ resources ─ resource_allocations → projects
               ├─ contractors ─ contractor_contracts → projects
               ├─ workers ─ worker_shifts → projects
               ├─ stakeholders
               └─ communications → stakeholders, projects
```

---

## 9. Security & Roles

| Role | Access |
|------|--------|
| owner | Full org |
| manager | All modules, no billing settings |
| engineer | Projects, materials, docs, charts |
| site | Materials, workers, site logs |
| viewer | Read-only |

UsersPanel permission: `morph_engi_access` (Admin bypass). Dev mode: any authenticated user.

Audit (future): who changed budget/contracts.

---

## 10. Development Phases

### Phase 1 — Foundation (this implementation)

- Auth, org, dashboard
- Projects CRUD + tasks + site logs
- All module CRUD APIs
- Svelte UI all sections
- AI assistant full tool suite

### Phase 2 — Depth

- Gantt-lite timeline view
- File upload to S3/MinIO
- BOQ import CSV
- PDF drawing preview

### Phase 3 — Integrations

- BIM 360 / ACC links
- MorphData asset sync
- GIS map pin for projects
- Mobile PWA offline site log

---

## 11. Folder Structure

```text
morph-engi/
├── design.md
├── README.md
├── .env.example
├── backend/
│   ├── Cargo.toml
│   └── src/
│       ├── main.rs
│       ├── config.rs
│       ├── db/
│       ├── middleware/
│       ├── services/
│       └── api/
└── frontend/
    ├── src/
    │   ├── App.svelte
    │   ├── app.css
    │   ├── lib/
    │   └── components/
    └── public/
```

---

## 12. Success Metrics

- Time to first project < 2 minutes
- Common daily task (site log + material) ≤ 3 clicks
- AI completes same task without UI navigation
- Page load < 1s on local dev
