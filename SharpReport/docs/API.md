# DataPulse API Documentation

## Authentication

### Login
```
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "admin@example.com",
  "password": "securepassword"
}

Response:
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Get Current User
```
GET /api/v1/auth/me
Authorization: Bearer <token>

Response:
{
  "id": "user-id",
  "email": "admin@example.com",
  "name": "Admin User",
  "role": "admin",
  "avatar_url": null
}
```

## Setup

### Check Setup Status
```
GET /api/v1/setup/status

Response:
{
  "is_completed": false,
  "metabase_initialized": false,
  "admin_user_created": false,
  "metabase_port": 3001
}
```

### Initialize Metabase
```
POST /api/v1/setup/initialize
Content-Type: application/json

{
  "jvm_path": "/path/to/java"
}

Response:
{
  "status": "initialized",
  "metabase_port": 3001
}
```

## Database Connections

### List Connections
```
GET /api/v1/databases
Authorization: Bearer <token>

Response:
[
  {
    "id": "conn-id",
    "name": "Production PostgreSQL",
    "engine": "postgres",
    "host": "db.example.com",
    "port": 5432,
    "database_name": "myapp_production",
    "username": "admin",
    "ssl_enabled": true
  }
]
```

### Create Connection
```
POST /api/v1/databases
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Production PostgreSQL",
  "engine": "postgres",
  "host": "db.example.com",
  "port": 5432,
  "database_name": "myapp_production",
  "username": "admin",
  "password": "securepassword",
  "ssl_enabled": true
}

Response:
{
  "id": "new-conn-id",
  "name": "Production PostgreSQL",
  "engine": "postgres",
  "host": "db.example.com",
  "port": 5432,
  "database_name": "myapp_production",
  "username": "admin",
  "ssl_enabled": true
}
```

## Dashboards

### List Dashboards
```
GET /api/v1/dashboards
Authorization: Bearer <token>

Response:
[
  {
    "id": "dash-id",
    "name": "Sales Overview",
    "description": "Monthly sales performance",
    "metabase_id": 1,
    "database_id": "conn-id",
    "is_public": false
  }
]
```

### Get Embed URL
```
GET /api/v1/dashboards/{id}/embed
Authorization: Bearer <token>

Response:
{
  "embed_url": "/embed/dashboard/{id}",
  "token": "signed-jwt-token"
}
```

## Embedding

### Generate Dashboard Embed Token
```
POST /api/v1/embed/dashboard/{id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "params": {
    "bordered": true,
    "titled": true,
    "theme": "dark"
  }
}

Response:
{
  "embed_url": "/embed/dashboard/{id}",
  "token": "signed-jwt-token",
  "expires_in": 600
}
```

## Metabase Proxy

All Metabase API endpoints are available under `/metabase/*`:

```
GET /metabase/api/database
Authorization: Bearer <token>
```

This will proxy the request to the Metabase instance running on the configured port.

## Error Responses

All errors follow this format:

```json
{
  "error": "Error message"
}
```

With appropriate HTTP status codes:
- `400` Bad Request
- `401` Unauthorized
- `403` Forbidden
- `404` Not Found
- `500` Internal Server Error