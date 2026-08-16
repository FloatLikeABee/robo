# Academi Backend

A Go-based backend API for the Academi AI study assistant application.

## Features

- RESTful API with Gin framework
- AI provider integration (OpenAI, Anthropic, Azure OpenAI, Ollama)
- JWT authentication
- BadgerDB for data persistence
- CORS support for frontend integration
- Multiple services: Auth, AI Chat, Community, Docs, Guides, Notifications

## Prerequisites

- Go 1.21 or higher
- Access to at least one AI provider (OpenAI, Anthropic, Azure, or Ollama)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/FloatLikeABee/academi.git
cd academi/backend
```

2. Install dependencies:
```bash
go mod download
```

3. Create environment configuration:
```bash
cp .env.example .env
```

4. Edit `.env` with your configuration:
```bash
# Server
SERVER_PORT=8080
SERVER_MODE=debug

# Database
DB_PATH=./data/badger

# JWT
JWT_SECRET=change-this-to-a-secure-random-string

# AI Provider Configuration
AI_DEFAULT_PROVIDER=openai

# OpenAI
OPENAI_API_KEY=your-openai-api-key-here
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_MODEL=gpt-4

# Anthropic (Claude)
ANTHROPIC_API_KEY=your-anthropic-api-key-here
ANTHROPIC_BASE_URL=https://api.anthropic.com/v1
ANTHROPIC_MODEL=claude-3-opus-20240229

# Azure OpenAI
AZURE_OPENAI_API_KEY=your-azure-openai-api-key-here
AZURE_OPENAI_BASE_URL=https://your-resource.openai.azure.com
AZURE_OPENAI_MODEL=gpt-4

# Ollama (Local)
OLLAMA_API_KEY=
OLLAMA_BASE_URL=http://localhost:11434/v1
OLLAMA_MODEL=llama2
```

## Running the Backend

### Development Mode

```bash
go run cmd/main.go
```

### Production Build

```bash
go build -o academi-backend cmd/main.go
./academi-backend
```

### Windows

```bash
go build -o academi-backend.exe cmd/main.go
.\academi-backend.exe
```

## API Endpoints

### Health Check
```
GET /health
```

### Authentication
```
POST /api/v1/auth/register
POST /api/v1/auth/login
```

### AI Services
```
GET  /api/v1/ai/providers        # List available AI providers (public)
POST /api/v1/ai/chat             # Send chat message (requires auth)
POST /api/v1/ai/summarize        # Summarize content (requires auth)
POST /api/v1/ai/generate-guide   # Generate study guide (requires auth)
POST /api/v1/ai/moderate         # Moderate content (requires auth)
```

### Community
```
GET    /api/v1/community/posts
POST   /api/v1/community/posts
GET    /api/v1/community/posts/:id
```

### Documentation
```
GET    /api/v1/docs/docs
POST   /api/v1/docs/docs
GET    /api/v1/docs/docs/:id
```

### Guides
```
GET    /api/v1/guides
POST   /api/v1/guides
GET    /api/v1/guides/:id
```

### Notifications
```
GET    /api/v1/notifications
POST   /api/v1/notifications
```

## AI Provider Configuration

The backend supports multiple AI providers. Configure your preferred provider in the `.env` file:

### OpenAI
- Set `AI_DEFAULT_PROVIDER=openai`
- Provide `OPENAI_API_KEY`
- Optionally customize `OPENAI_BASE_URL` and `OPENAI_MODEL`

### Anthropic (Claude)
- Set `AI_DEFAULT_PROVIDER=anthropic`
- Provide `ANTHROPIC_API_KEY`
- Optionally customize `ANTHROPIC_BASE_URL` and `ANTHROPIC_MODEL`

### Azure OpenAI
- Set `AI_DEFAULT_PROVIDER=azure`
- Provide `AZURE_OPENAI_API_KEY` and `AZURE_OPENAI_BASE_URL`
- Set `AZURE_OPENAI_MODEL` to your deployment name

### Ollama (Local)
- Set `AI_DEFAULT_PROVIDER=ollama`
- Ensure Ollama is running locally on port 11434
- Optionally customize `OLLAMA_BASE_URL` and `OLLAMA_MODEL`

## CORS Configuration

By default, the backend allows requests from:
- `http://localhost:3000`
- `http://localhost:8081`

To add more origins, modify the `CORS.Origins` in `internal/config/config.go`.

## Database

The backend uses BadgerDB (an embedded key-value store). Data is stored in the path specified by `DB_PATH` in the `.env` file (default: `./data/badger`).

## Deployment

### Docker (Recommended)

Create a `Dockerfile`:

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o academi-backend cmd/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/academi-backend .
COPY --from=builder /app/.env.example .env
EXPOSE 8080
CMD ["./academi-backend"]
```

Build and run:
```bash
docker build -t academi-backend .
docker run -p 8080:8080 --env-file .env academi-backend
```

### Systemd Service (Linux)

Create `/etc/systemd/system/academi-backend.service`:

```ini
[Unit]
Description=Academi Backend API
After=network.target

[Service]
Type=simple
User=academi
WorkingDirectory=/opt/academi/backend
ExecStart=/opt/academi/backend/academi-backend
Restart=always
RestartSec=5
EnvironmentFile=/opt/academi/backend/.env

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl enable academi-backend
sudo systemctl start academi-backend
```

### Windows Service

Use NSSM (Non-Sucking Service Manager):
```bash
nssm install AcademiBackend "C:\path\to\academi-backend.exe"
nssm set AcademiBackend AppDirectory "C:\path\to\backend"
nssm set AcademiBackend AppEnvironmentExtra "SERVER_PORT=8080"
nssm start AcademiBackend
```

## Security Considerations

1. **JWT Secret**: Always use a strong, random JWT secret in production
2. **API Keys**: Never commit `.env` file with real API keys
3. **CORS**: Restrict CORS origins to your actual frontend domains
4. **Database**: Ensure proper file permissions on the database directory
5. **HTTPS**: Use a reverse proxy (nginx, Caddy) for SSL/TLS in production

## Development

### Project Structure
```
backend/
├── cmd/
│   └── main.go           # Application entry point
├── internal/
│   ├── ai/               # AI service
│   ├── auth/             # Authentication service
│   ├── community/        # Community features
│   ├── config/           # Configuration management
│   ├── database/         # Database operations
│   ├── docs/             # Documentation service
│   ├── guide/            # Study guides
│   ├── models/           # Data models
│   ├── notifications/    # Notification service
│   └── middleware/       # HTTP middleware
├── api/                  # API route definitions (if needed)
├── migrations/           # Database migrations
├── pkg/                  # Public packages
├── .env.example          # Environment template
├── go.mod                # Go module definition
└── go.sum                # Go module checksums
```

### Adding New Services

1. Create service in `internal/your-service/`
2. Implement service with `RegisterRoutes` method
3. Register in `cmd/main.go`
4. Add models in `internal/models/` if needed

## Troubleshooting

### Port Already in Use
Change `SERVER_PORT` in `.env` or stop the conflicting process.

### Database Lock Error
Ensure only one instance of the backend is running, or check for stale lock files in the database directory.

### AI Provider Errors
- Verify API keys are correct
- Check API provider status
- Ensure network connectivity to the API endpoint
- Review rate limits for your API plan

### CORS Issues
Add your frontend domain to the `CORS.Origins` list in the configuration.

## License

[Your License Here]

## Support

For issues and questions, please open an issue on the GitHub repository.
