# Academi Ledger — Account Booking & Warehouse Platform

## 1. Product Vision

A modern accounting + booking + warehouse management platform designed as a microservice ecosystem.

The goal is to:
- Make bookkeeping approachable for small businesses.
- Support asset imports and financial migration from existing systems.
- Provide lightweight warehouse management.
- Support mobile-first operations.
- Offer a visually premium UI with strong UX.
- Be extensible into a larger platform ecosystem.

This system is not intended to replace enterprise ERP systems like SAP initially.
Instead, it focuses on:
- SMEs
- Schools
- Retail shops
- Small warehouses
- Agencies
- Internal operations teams

---

# 2. Core Philosophy

## Key Principles

### Simplicity First
Most accounting systems feel old, crowded, and intimidating.
This system should feel:
- Clean
- Fast
- Friendly
- Modern
- Guided

### Platform Ready
Every module should work independently as a microservice.

### Import Friendly
Businesses already have data.
Migration support is critical.

### Mobile Optimized
Many users operate inventory and bookkeeping from phones.

### UX Over Complexity
Reduce accounting fear by:
- Human-readable wording
- Guided workflows
- Visual summaries
- Smart defaults
- Validation helpers

---

# 3. Branding & Visual Direction

## Visual Style
A premium dashboard aesthetic.

### Primary Colors
| Purpose | Color |
|---|---|
| Deep Violet | #5B3FD6 |
| Dark Grey | #1E1E24 |
| Accent Yellow | #F5C542 |
| Surface Dark | #2B2B33 |
| Surface Light | #F9F9FC |

---

## Theme System

### Dark Theme
Primary experience.
Feels cinematic and professional.

- Deep violet gradients
- Glassmorphism panels
- Soft shadows
- Slight neon highlights
- Yellow action accents

### Light Theme
Minimal, modern SaaS aesthetic.

---

## UI Mood
Inspired by:
- Linear
- Notion
- Stripe
- Vercel
- modern fintech dashboards

---

# 4. High Level Architecture

```text
                ┌────────────────────┐
                │   React Frontend   │
                │  Web + Mobile PWA  │
                └─────────┬──────────┘
                          │
                   API Gateway
                          │
 ┌──────────────────────────────────────────────────┐
 │                Backend Services                  │
 └──────────────────────────────────────────────────┘

  ┌──────────────┐
  │ Auth Service │
  └──────────────┘

  ┌───────────────────┐
  │ Accounting Service│
  └───────────────────┘

  ┌──────────────────┐
  │ Booking Service  │
  └──────────────────┘

  ┌──────────────────┐
  │ Warehouse Service│
  └──────────────────┘

  ┌──────────────────┐
  │ Asset Service    │
  └──────────────────┘

  ┌──────────────────┐
  │ Import Service   │
  └──────────────────┘

  ┌──────────────────┐
  │ Notification Svc │
  └──────────────────┘

            │
      MySQL Database
            │
      Redis Cache Layer
```

---

# 5. Technology Stack

## Frontend

### Framework
- ReactJS
- TypeScript

### Styling
- TailwindCSS
- Framer Motion
- ShadCN/UI

### State Management
- Zustand

### Mobile Support
- Responsive web app
- PWA support
- Future React Native compatibility

### Charts
- Recharts

---

## Backend

### Language
- Golang

### Frameworks
- Fiber OR Gin

Recommendation:
- Gin for maturity
- Fiber for speed and modern DX

---

## Database

### Primary
- MySQL

### Cache
- Redis

---

## Authentication
- JWT
- Refresh Tokens
- RBAC
- Multi-tenant organization support

---

## File Storage
- S3-compatible storage
- MinIO locally

---

## Infrastructure

### Containerization
- Docker
- Docker Compose

### Future Scale
- Kubernetes

---

# 6. Frontend Design System

## Layout Structure

### Sidebar
Contains:
- Dashboard
- Bookings
- Accounting
- Warehouse
- Assets
- Imports
- Reports
- Settings

### Top Navigation
Contains:
- Organization switcher
- Search
- Notifications
- Theme switch
- Profile

---

## UI Components

### Cards
Rounded:
- 20px radius
- soft shadow
- glass blur in dark mode

### Tables
Features:
- Sticky headers
- Sorting
- Pagination
- Column visibility
- Export buttons

### Forms
Features:
- Smart validation
- Inline hints
- Auto-save drafts

### Animations
Use subtle:
- fade
- slide
- scale
- hover transitions

Avoid excessive motion.

---

# 7. Core Modules

# 7.1 Authentication & Organization

## Features
- Login
- Register
- Invite users
- Organization creation
- Role permissions
- Team management

---

## Roles

### Owner
Full access.

### Accountant
Financial management.

### Warehouse Staff
Inventory access only.

### Viewer
Read-only.

### Admin
Organization management.

---

# 7.2 Booking Module

This is the operational booking/account entry system.

## Features
- Create bookings
- Recurring bookings
- Attach invoices
- Attach receipts
- Booking approval workflow
- Draft bookings
- Booking status

---

## Booking States
| State | Meaning |
|---|---|
| Draft | Not finalized |
| Pending | Awaiting approval |
| Approved | Accepted |
| Posted | Sent to accounting ledger |
| Cancelled | Invalidated |

---

## Booking Data Model

### Fields
- booking_id
- customer_id
- booking_date
- due_date
- subtotal
- tax
- total
- currency
- status
- notes
- attachments
- created_by

---

# 7.3 Accounting Module

This is the financial engine.

---

## Accounting Concepts Needed

### Chart of Accounts
The system must include:

#### Asset Accounts
- Cash
- Bank
- Inventory
- Equipment
- Accounts Receivable

#### Liability Accounts
- Loans
- Accounts Payable
- Tax Payable

#### Equity Accounts
- Owner Equity
- Retained Earnings

#### Revenue Accounts
- Sales Revenue
- Service Revenue

#### Expense Accounts
- Utilities
- Salary
- Rent
- Office Expense

---

## Double Entry Accounting

Every transaction must:
- Debit one account
- Credit another account

Example:

| Account | Debit | Credit |
|---|---|---|
| Cash | 1000 | |
| Sales Revenue | | 1000 |

---

## General Ledger

The ledger records all accounting transactions.

---

## Journal Entries

### Structure
- Entry number
- Date
- Reference
- Account lines
- Debit amount
- Credit amount
- Notes

---

## Fiscal Configuration

Users must configure:
- Country
- Fiscal year start
- Tax system
- Base currency
- Tax percentages
- Accounting method

---

## Opening Balances

VERY IMPORTANT.

Businesses already have existing financials.

The platform must support:
- Opening balances
- Existing assets
- Existing liabilities
- Existing customer debts
- Existing supplier debts

---

## Accounting Setup Wizard

On first launch:

### Step 1
Organization info.

### Step 2
Country & accounting region.

### Step 3
Fiscal year setup.

### Step 4
Opening balances.

### Step 5
Import historical data.

### Step 6
Warehouse setup.

---

# 7.4 Warehouse Management

This is lightweight WMS.

---

## Features

### Inventory Tracking
- SKU tracking
- Barcode support
- Batch tracking
- Quantity tracking
- Unit tracking

---

## Warehouse Operations

### Stock In
Receiving items.

### Stock Out
Dispatching items.

### Transfer
Warehouse to warehouse.

### Adjustment
Correct inventory counts.

---

## Product Structure

### Fields
- SKU
- Name
- Category
- Cost price
- Selling price
- Quantity
- Unit
- Barcode
- Supplier
- Warehouse
- Reorder threshold

---

## Inventory Valuation

Support:
- FIFO
- Weighted Average

---

## Warehouse Dashboard

### Metrics
- Low stock
- Fast-moving items
- Dead stock
- Inventory value
- Incoming stock

---

# 7.5 Asset Management

Tracks company assets.

---

## Asset Types
- Equipment
- Furniture
- Vehicles
- Technology
- Buildings

---

## Asset Features
- Asset registration
- Asset depreciation
- Asset location
- Asset condition
- Asset assignment
- Asset lifecycle

---

## Depreciation Methods
- Straight line
- Declining balance

---

## Asset Status
| Status | Meaning |
|---|---|
| Active | In use |
| Maintenance | Under repair |
| Retired | No longer used |
| Sold | Disposed |

---

# 7.6 Import & Integration Service

One of the most important modules.

---

## Supported Imports
- CSV
- JSON
- Excel

---

## Importable Data
- Customers
- Suppliers
- Products
- Assets
- Transactions
- Journal entries
- Opening balances

---

## HTTP Import Connector

Allow importing from external APIs.

### Example
```json
{
  "url": "https://example.com/api/products",
  "method": "GET",
  "headers": {
    "Authorization": "Bearer token"
  }
}
```

---

## Import Flow

### Step 1
Upload/import source.

### Step 2
Preview parsed data.

### Step 3
Map columns.

### Step 4
Validate.

### Step 5
Import.

### Step 6
Generate logs.

---

## Validation Rules
- Duplicate detection
- Missing required fields
- Invalid account references
- Invalid SKUs
- Currency mismatch

---

# 8. Reports & Analytics

## Financial Reports

### Reports
- Profit & Loss
- Balance Sheet
- Cash Flow
- Trial Balance
- Tax Summary

---

## Warehouse Reports
- Inventory valuation
- Low stock
- Movement history
- Warehouse performance

---

## Dashboard Widgets

### Finance
- Monthly revenue
- Expenses
- Net profit
- Cash position

### Warehouse
- Total stock
- Inventory value
- Fast-moving items

### Booking
- Pending bookings
- Approval rate

---

# 9. Database Design

# Core Tables

## Users
```sql
users
- id
- organization_id
- name
- email
- password_hash
- role
- created_at
```

---

## Organizations
```sql
organizations
- id
- name
- country
- currency
- fiscal_year_start
- created_at
```

---

## Accounts
```sql
accounts
- id
- organization_id
- code
- name
- type
- parent_id
```

---

## Journal Entries
```sql
journal_entries
- id
- organization_id
- reference
- date
- description
```

---

## Journal Lines
```sql
journal_lines
- id
- journal_entry_id
- account_id
- debit
- credit
```

---

## Products
```sql
products
- id
- sku
- name
- quantity
- cost_price
- selling_price
```

---

## Inventory Movements
```sql
inventory_movements
- id
- product_id
- warehouse_id
- type
- quantity
- reference
```

---

## Assets
```sql
assets
- id
- asset_tag
- name
- value
- depreciation_method
- status
```

---

# 10. Backend Service Design

# API Gateway

Responsibilities:
- Route requests
- JWT validation
- Rate limiting
- Logging
- API aggregation

---

# Auth Service

Endpoints:
```text
POST /auth/login
POST /auth/register
POST /auth/refresh
POST /auth/logout
```

---

# Accounting Service

Endpoints:
```text
GET /accounts
POST /journal-entries
GET /ledger
GET /reports/profit-loss
```

---

# Warehouse Service

Endpoints:
```text
GET /products
POST /stock-in
POST /stock-out
POST /transfers
```

---

# Asset Service

Endpoints:
```text
GET /assets
POST /assets
POST /assets/import
```

---

# Import Service

Endpoints:
```text
POST /imports/csv
POST /imports/json
POST /imports/http
```

---

# 11. Environment Variables

```env
# DATABASE
DB_HOST=
DB_PORT=
DB_USER=
DB_PASSWORD=
DB_NAME=

# REDIS
REDIS_HOST=
REDIS_PORT=

# JWT
JWT_SECRET=
JWT_EXPIRY=

# STORAGE
S3_ENDPOINT=
S3_ACCESS_KEY=
S3_SECRET_KEY=
S3_BUCKET=

# APP
APP_ENV=
APP_PORT=
```

---

# 12. Security Design

## Requirements
- Password hashing
- JWT authentication
- Refresh token rotation
- Role permissions
- Input sanitization
- Rate limiting
- Audit logs
- File scanning

---

# Audit Logs
Track:
- Login activity
- Financial edits
- Inventory changes
- Imports
- User permission changes

---

# 13. UX Flows

# First Time User Flow

```text
Signup
 → Create Organization
 → Configure Country
 → Configure Fiscal Year
 → Import Existing Data
 → Setup Warehouse
 → Dashboard Ready
```

---

# Booking Flow

```text
Create Booking
 → Add Items
 → Select Accounts
 → Save Draft
 → Approval
 → Post To Ledger
```

---

# Inventory Flow

```text
Receive Stock
 → Validate SKU
 → Update Quantity
 → Generate Inventory Movement
 → Update Accounting Inventory Value
```

---

# 14. Responsive Design

## Desktop
- Full sidebar
- Multi-column dashboards
- Advanced tables

---

## Tablet
- Collapsible sidebar
- Reduced widgets

---

## Mobile
- Bottom navigation
- Card-based layout
- Swipe interactions
- Quick actions

---

# 15. Future Expansion

## Possible Future Modules
- POS system
- Payroll
- HR management
- CRM
- AI forecasting
- AI accounting assistant
- OCR invoice scanning
- Banking integration
- Mobile native app
- Supplier portal
- Purchase orders
- E-commerce integration

---

# 16. Recommended Folder Structure

# Frontend

```text
src/
 ├── components/
 ├── modules/
 ├── pages/
 ├── layouts/
 ├── hooks/
 ├── services/
 ├── store/
 ├── theme/
 ├── routes/
 └── utils/
```

---

# Backend

```text
backend/
 ├── cmd/
 ├── internal/
 │    ├── auth/
 │    ├── accounting/
 │    ├── warehouse/
 │    ├── booking/
 │    ├── assets/
 │    ├── imports/
 │    └── reports/
 ├── pkg/
 ├── configs/
 ├── migrations/
 └── docker/
```

---

# 17. Recommended Development Phases

# Phase 1 — Foundation
- Authentication
- Organization setup
- Theme system
- Dashboard shell
- RBAC

---

# Phase 2 — Accounting Core
- Chart of accounts
- Journal entries
- Ledger
- Reports
- Opening balances

---

# Phase 3 — Warehouse
- Products
- Inventory tracking
- Movements
- Warehouse reports

---

# Phase 4 — Asset Management
- Asset registration
- Depreciation
- Lifecycle management

---

# Phase 5 — Import Engine
- CSV import
- JSON import
- HTTP import
- Validation engine

---

# Phase 6 — Polish
- Mobile optimization
- Advanced analytics
- Notification system
- UX improvements
- AI assistant

---

# 18. Important Real-World Considerations

## Accounting is Country-Specific
Taxes and fiscal logic vary.

The architecture should support:
- Country adapters
- Tax plugins
- Regional formatting

---

## Data Integrity is Critical
Never allow:
- Unbalanced journal entries
- Negative stock unless enabled
- Deleted financial history

Use:
- Soft deletes
- Audit trails
- Immutable accounting records

---

## Inventory + Accounting Must Sync
Inventory movements should affect:
- Inventory asset account
- Cost of goods sold
- Revenue tracking

---

## Imports Need Error Recovery
Large imports fail often.

Need:
- Retry support
- Partial import recovery
- Validation previews
- Import history

---

# 19. Final Product Feel

The final product should feel like:

> "A modern financial operating system, not an old accounting tool."

It should:
- Feel elegant
- Reduce operational stress
- Make accounting understandable
- Work beautifully on mobile
- Scale into a larger business ecosystem
- Feel fast and premium

The biggest differentiator should be:
- UX quality
- simplicity
- integrations
- visual clarity
- workflow smoothness
- modern design language
- mobile friendliness

