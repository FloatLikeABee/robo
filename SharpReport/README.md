# DataPulse - Metabase-Powered Analytics Platform

A modern analytics platform that embeds Metabase as its reporting engine, providing a sleek, intuitive interface for database connectivity, data manipulation, and visualization.

## Features

- **Metabase Integration**: Embed Metabase dashboards and queries with signed JWT tokens
- **Database Connectivity**: Connect to any Metabase-compatible database
- **First-Run Setup**: Automatic Metabase initialization and configuration
- **Modern UI**: Svelte 5 frontend with TailwindCSS and dark/light themes
- **Rust Backend**: Fast, reliable, and secure API server
- **Docker Deployment**: Easy containerized deployment

## Tech Stack

### Backend
- Rust 2024 Edition
- Axum web framework
- SQLx for database access
- Tokio for async runtime
- Metabase process management

### Frontend
- Svelte 5 with Runes
- TailwindCSS v4
- Lucide icons
- ECharts for visualization

## Getting Started

### Prerequisites
- Rust 1.84+
- Node.js 22+
- Java 17+ (for Metabase)
- Docker (optional)

### Development Setup

1. Clone the repository:
```bash
git clone https://github.com/your-repo/datapulse.git
cd datapulse
```

2. Set up environment variables:
```bash
cp .env.example .env
# Edit .env with your settings
```

3. Start the development servers:
```bash
# In one terminal
cd backend
cargo run

# In another terminal
cd frontend
npm install
npm run dev
```

### Docker Deployment

1. Build and start containers:
```bash
cd deploy
docker-compose up --build
```

2. Access the application at `http://localhost:3000`

## Project Structure

```
datapulse/
├── backend/          # Rust backend
├── frontend/         # Svelte frontend
├── deploy/           # Docker configuration
├── metabase/         # Metabase JAR storage
├── docs/             # Documentation
└── design.md         # Architecture design
```

## Configuration

See `.env.example` for all configuration options.

## AI provider configuration

SharpReport (DataPulse) uses **MorphAI** ([`pkg/morphai-rs`](../../pkg/morphai-rs)) for the report-building assistant in the UI (SQL help, chart guidance, Metabase context).

### 1. Get a DashScope API key

Create an API key at [Alibaba Cloud DashScope](https://dashscope.aliyun.com/) (Qwen models).

### 2. Configure the backend

Copy and edit the repo **`.env`**:

```bash
cp .env.example .env
```

Add:

```bash
MORPH_AI_API_KEY=sk-your-dashscope-key
MORPH_AI_MODEL=qwen3-max
# optional:
# MORPH_AI_API_URL=https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation
```

| Variable | Default | Purpose |
|----------|---------|---------|
| `MORPH_AI_API_KEY` | _(empty)_ | **Required** for LLM assistant |
| `MORPH_AI_MODEL` | `qwen3-max` | Chat model |
| `MORPH_AI_API_URL` | DashScope text-generation URL | Optional endpoint override |

**Legacy fallbacks:** `GEMINI_API_KEY`, `GEMINI_MODEL`.

The Rust backend loads `.env` from the `SharpReport/` directory when you run `cargo run` from `backend/`.

### 3. Restart

```bash
cd backend && cargo run
```

Or: `./start-all.sh restart sharpreport-api`

### 4. Verify

1. Open `http://localhost:5178` and sign in (UsersPanel account).
2. Open the **AI assistant** drawer in Report Builder or Reports hub.
3. Ask about SQL or report setup — without a key you get static guidance mentioning `MORPH_AI_API_KEY`.

**Assistant API:** `POST /api/v1/assistant/chat` (also aliased as `/api/v1/reports/assistant/chat`).

See also: [`AI_ASSISTANT_MORPHAI_CONTRACT.md`](../../AI_ASSISTANT_MORPHAI_CONTRACT.md).

## License

Proprietary - All rights reserved

## Roadmap

- [ ] Complete backend API implementation
- [ ] Finish frontend components
- [ ] Implement Metabase embedding
- [ ] Add database management UI
- [ ] Build query builder interface
- [ ] Create dashboard designer
- [ ] Implement user management
- [ ] Add SSO integration
- [ ] Production hardening

## Contributing

This is a proprietary project. Contributions are not currently accepted.

## Support

For issues or questions, please contact support@datapulse.app