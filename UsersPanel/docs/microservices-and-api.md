# API usage and microservice integration

This document describes how to work with the **Users Panel** HTTP API from any client, and how to position it in a **microservice** architecture.

## OpenAPI (Swagger)

The backend serves a complete **OpenAPI 3** description and a browser UI (Swagger UI) from the same process and port as the API (`HOST` and `PORT`, default `http://127.0.0.1:5001`).

| Resource | URL (default dev) |
|----------|-------------------|
| **Swagger UI** | `http://127.0.0.1:5001/swagger-ui` |
| **OpenAPI JSON** | `http://127.0.0.1:5001/openapi.json` |

With the Vite dev server, these paths are proxied so you can use:

- `http://localhost:5173/swagger-ui`
- `http://localhost:5173/openapi.json`

Use the JSON in **Postman**, **Insomnia**, **OpenAPI Generator** (TypeScript, Kotlin, Go, etc.), or import into an API gateway that supports OpenAPI 3.

For protected routes, in Swagger click **Authorize**, enter `Bearer <your_jwt>`, and call admin endpoints with an account that has the **Admin** role.

---

## Authentication model

- **Session:** JWT in the `Authorization: Bearer <token>` header (except register, login, verify, forgot/reset, OAuth start/callback).
- **Claims (subject):** user id, email, username, `roles` (string array; role **display names** as stored in `plat_roles`, e.g. `Admin`, `Main Panel`), and a `defaultChannelId` string used as a tenant or channel id in integrations.
- **Errors:** failed requests return JSON `{"error": "..."}`. An invalid or missing token yields **401**.

There is no separate “API key” in this project; services should either reuse JWT validation (below) or call this service when they need a trusted user profile or permission list.

---

## Microservice deployment patterns

### 1. Dedicated identity and policy service (recommended)

- Run **this API** (and its MySQL database) as the **source of truth** for users, roles, and permissions.
- **Edge / BFF** or **API gateway** forwards `Authorization: Bearer` to each downstream service.
- **Other services** that need authorization have two main options:
  1. **Shared JWT secret:** configure the same `JWT_SECRET` (and the same claim shape) in each service, validate the JWT locally, and trust `roles` / `sub` in the token. This is simple and fast; rotation of `JWT_SECRET` must be coordinated.
  2. **Introspection / userinfo:** the downstream service only forwards the token to this service (e.g. `GET /api/auth/user` and/or `GET /api/auth/permissions`) when it must resolve a user. Higher latency, but no shared secret in every service.

“Permissions” in the token sense are a **denormalized list of slugs** (see `GET /api/auth/permissions`). Migrations and `permissions_for_roles` define how role names map to those slugs; the **admin** API lets you add permission rows and link them to roles in `plat_role_permissions`.

### 2. Monolith with future extraction

- Start with one gateway that routes `/api` to this service and serves the admin SPA. When you add new services, move routes behind the same gateway and keep a single public origin; set **`FRONTEND_ORIGIN`** to that origin for OAuth and email/reset links.

### 3. Network and CORS

- CORS in this project allows **`FRONTEND_ORIGIN`** (default `http://localhost:5173`). For additional browser origins, extend the CORS layer in the backend (or put the browser behind a reverse proxy to one origin). Server-to-server calls do not need CORS.

---

## End-to-end: users, roles, and permissions

The API does **not** provide “admin create user with password” as a single call. New password users are created with **`POST /api/auth/register`** and email verification, or you can use **`BOOTSTRAP_ADMIN_EMAIL` / `BOOTSTRAP_ADMIN_PASSWORD`** on first boot for a seed admin, or **Google OAuth**. Admins can list users and **assign roles**; they do not create new rows without register/bootstrap/OAuth (unless you add an endpoint later).

### A. Create an end user (self-service)

1. **`POST /api/auth/register`** with JSON `{ "email", "username", "password" }` (password ≥ 8 characters).
2. In development, the verification link is **logged** by the server (`/api/auth/verify-email?token=...`). Open it.
3. **`POST /api/auth/login`** with `{ "email", "password" }` → response includes `token` and `user`.

### B. Get an admin JWT

- Sign in with a user whose `roles` include **`Admin`**, or use the bootstrap user from env.

### C. Add a new **role** (name is a display name, e.g. `Support Tier 1`)

```http
POST /api/admin/roles
Authorization: Bearer <admin_jwt>
Content-Type: application/json

{ "name": "Support Tier 1", "description": "optional" }
```

List roles: **`GET /api/admin/roles`**.

### D. Add a **permission** (slug, lowercase, `a_z0-9_`)

```http
POST /api/admin/permissions
Authorization: Bearer <admin_jwt>
Content-Type: application/json

{ "name": "support_ticket_read", "description": "Read support tickets" }
```

The OpenAPI spec documents exact validation rules.

### E. Attach permissions to a role (replace the full set for that role)

1. Get permission **ids** from **`GET /api/admin/permissions`** (or the overview).
2. Call **`PUT /api/admin/roles/{role_name}/permissions`** with the role’s **URL-encoded** display name, e.g. for `Main Panel`:

`PUT /api/admin/roles/Main%20Panel/permissions`

```json
{ "permission_ids": ["<uuid-1>", "<uuid-2>"] }
```

The request field name is **`permission_ids`**, matching the OpenAPI / Swagger `PutRolePermissionsBody` schema.

3. The overview **`GET /api/admin/permissions/overview`** returns `roleNames`, `assignments` (map from role name to permission id list), and the permission catalog.

### F. Give a user a role (assign by role name)

1. List users: **`GET /api/admin/users`** and copy the user’s `id`.
2. **`PATCH /api/admin/users/{user_id}/roles`**

```json
{ "roles": ["Admin", "Main Panel", "Support Tier 1"] }
```

Each string must be an existing role **name** in `plat_roles`.

### G. How a user’s permissions are resolved

- **JWT** carries `roles` as display names.
- **`GET /api/auth/permissions`** (authenticated) returns the **merged list of permission slugs** for those roles, based on `plat_role_permissions` and the static map in the backend (see `permissions_for_roles` in code for built-in slugs; custom permissions are stored in the DB and resolved through role links).

Use this in other services: either read slugs from a validated JWT plus your own service-specific checks, or call this API when you need a fresh permission list.

---

## Summary

| Goal | Primary endpoints |
|------|-------------------|
| OpenAPI for codegen / tools | `GET /openapi.json` |
| Register & login | `POST /api/auth/register`, `POST /api/auth/login` |
| Current user & permissions | `GET /api/auth/user`, `GET /api/auth/permissions` |
| Create role / permission | `POST /api/admin/roles`, `POST /api/admin/permissions` |
| Map permissions to a role | `PUT /api/admin/roles/{role_name}/permissions` |
| Assign roles to a user | `PATCH /api/admin/users/{user_id}/roles` |

For every path, method, body schema, and **Bearer** security, use the **Swagger** UI or the **OpenAPI** file as the source of truth.
